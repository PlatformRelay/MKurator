package controller

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// connectionReferencesSecret reports whether the connection authenticates or trusts via
// secretName. It matches the spoke's legacy credentialsSecretRef and TLS caSecretRef, plus any
// authentication-union member refs supplied by the caller.
//
// The union refs are passed in (not read here) because the reconciler/watch sees a
// down-converted v1alpha1 spoke whose authentication union was dropped; the caller re-reads the
// v1beta1 hub for them (mirroring the mqrest factory's resolveCredentialsSecretName, ADR-0027).
// Keeping this a pure function preserves its existing unit contract.
func connectionReferencesSecret(
	conn *messagingv1alpha1.QueueManagerConnection,
	secretName string,
	unionRefs ...string,
) bool {
	if conn.Spec.CredentialsSecretRef.Name == secretName {
		return true
	}
	if conn.Spec.TLS != nil && conn.Spec.TLS.CASecretRef != nil &&
		conn.Spec.TLS.CASecretRef.Name == secretName {
		return true
	}
	for _, ref := range unionRefs {
		if ref != "" && ref == secretName {
			return true
		}
	}
	return false
}

func requestsForSecret(
	ctx context.Context,
	c client.Client,
	secret *corev1.Secret,
) []reconcile.Request {
	logger := log.FromContext(ctx)
	connList := &messagingv1alpha1.QueueManagerConnectionList{}
	if err := c.List(ctx, connList, client.InNamespace(secret.Namespace)); err != nil {
		logger.Error(err, "list QueueManagerConnections for secret watch",
			"namespace", secret.Namespace, "secret", secret.Name)
		return nil
	}

	var reqs []reconcile.Request
	for i := range connList.Items {
		conn := &connList.Items[i]
		// The spoke dropped the authentication union on down-conversion; re-read the v1beta1 hub
		// so a union auth-Secret rotation enqueues its owner (AUTH-14, closing the AUTH-12 gap).
		unionRefs := unionSecretRefs(ctx, c, conn)
		if connectionReferencesSecret(conn, secret.Name, unionRefs...) {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{Namespace: secret.Namespace, Name: conn.Name},
			})
		}
	}
	return reqs
}

// unionSecretRefs returns the authentication-union member Secret names for conn by re-reading
// the v1beta1 hub (the spoke's union was dropped on down-conversion). It matches EVERY union
// member ref (basic/ltpa/clientCert) so later slices (AUTH-13/16) inherit the watch for free,
// even though only Basic is an accepted enum today.
//
// Degradation mirrors the mqrest factory (ADR-0027): if the hub is NotFound, the v1beta1 kind
// is not registered (older/fake clients), or any read error occurs, we return nil and fall back
// to spoke-only matching — a watch miss on a union secret is never worse than the pre-AUTH-14
// behaviour, and must not drop the enqueue for legacy refs.
func unionSecretRefs(
	ctx context.Context,
	c client.Client,
	conn *messagingv1alpha1.QueueManagerConnection,
) []string {
	hub := &messagingv1beta1.QueueManagerConnection{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: conn.Namespace, Name: conn.Name}, hub); err != nil {
		if !k8serrors.IsNotFound(err) && !meta.IsNoMatchError(err) && !runtime.IsNotRegisteredError(err) {
			log.FromContext(ctx).V(1).Info(
				"secret watch: re-read v1beta1 hub for authentication union failed; matching spoke refs only",
				"connection", conn.Name, "namespace", conn.Namespace, "error", err.Error())
		}
		return nil
	}
	a := hub.Spec.Authentication
	if a == nil {
		return nil
	}
	var refs []string
	if a.Basic != nil {
		refs = append(refs, a.Basic.SecretRef.Name)
	}
	if a.LTPA != nil {
		refs = append(refs, a.LTPA.SecretRef.Name)
	}
	if a.ClientCert != nil {
		refs = append(refs, a.ClientCert.SecretRef.Name)
	}
	return refs
}

func secretEnqueueMapper(c client.Client) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		secret, ok := obj.(*corev1.Secret)
		if !ok {
			return nil
		}
		return requestsForSecret(ctx, c, secret)
	}
}

func secretCredentialsStripped(secret *corev1.Secret) bool {
	return len(secret.Data) == 0 && len(secret.StringData) == 0
}

func secretContentChanged(oldSecret, newSecret *corev1.Secret) bool {
	dataChanged := !reflect.DeepEqual(oldSecret.Data, newSecret.Data) ||
		!reflect.DeepEqual(oldSecret.StringData, newSecret.StringData)
	if dataChanged {
		return true
	}
	// ResourceVersion catches rotations when the informer cache strips Secret data (ADR-0023 / ARCH-05).
	if secretCredentialsStripped(oldSecret) && secretCredentialsStripped(newSecret) &&
		oldSecret.ResourceVersion != "" && newSecret.ResourceVersion != "" &&
		oldSecret.ResourceVersion != newSecret.ResourceVersion {
		return true
	}
	return false
}

func secretWatchPredicates() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			_, ok := e.Object.(*corev1.Secret)
			return ok
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSecret, okOld := e.ObjectOld.(*corev1.Secret)
			newSecret, okNew := e.ObjectNew.(*corev1.Secret)
			if !okOld || !okNew {
				return false
			}
			return secretContentChanged(oldSecret, newSecret)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			_, ok := e.Object.(*corev1.Secret)
			return ok
		},
	}
}
