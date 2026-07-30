package controller

import (
	"context"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// connectionReferencesSecret reports whether conn authenticates or trusts via secretName. It
// matches the legacy credentialsSecretRef, the TLS caSecretRef, and every authentication-union
// member ref (basic/ltpa/clientCert) read natively from the v1beta1 spec — so every auth mode
// inherits the Secret watch for free (Basic/AUTH-12, LTPA/AUTH-13, ClientCert/AUTH-16) without
// any mode-specific code here.
//
// Every deref is nil-guarded: v1beta1 makes credentialsSecretRef, authentication, and each union
// member optional pointers, so a stale informer cache that briefly returns a partially-populated
// QMC is matched on whatever refs ARE readable rather than panicking.
func connectionReferencesSecret(conn *messagingv1beta1.QueueManagerConnection, secretName string) bool {
	if conn.Spec.CredentialsSecretRef != nil && conn.Spec.CredentialsSecretRef.Name == secretName {
		return true
	}
	if conn.Spec.TLS != nil && conn.Spec.TLS.CASecretRef != nil &&
		conn.Spec.TLS.CASecretRef.Name == secretName {
		return true
	}
	if a := conn.Spec.Authentication; a != nil {
		if a.Basic != nil && a.Basic.SecretRef.Name == secretName {
			return true
		}
		if a.LTPA != nil && a.LTPA.SecretRef.Name == secretName {
			return true
		}
		if a.ClientCert != nil && a.ClientCert.SecretRef.Name == secretName {
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
	connList := &messagingv1beta1.QueueManagerConnectionList{}
	if err := c.List(ctx, connList, client.InNamespace(secret.Namespace)); err != nil {
		logger.Error(err, "list QueueManagerConnections for secret watch",
			"namespace", secret.Namespace, "secret", secret.Name)
		return nil
	}

	var reqs []reconcile.Request
	for i := range connList.Items {
		conn := &connList.Items[i]
		if conn.Spec.Authentication == nil {
			// A v1beta1 QMC may legitimately omit the authentication union (defaults to Basic
			// against credentialsSecretRef), and a stale informer read can briefly return one
			// with the union not yet populated. Either way the mapper degrades gracefully:
			// connectionReferencesSecret nil-guards every deref and matches whatever refs ARE
			// readable. Logged at V(1) so the fallback is observable without spamming the
			// default log for every legacy-Basic connection.
			logger.V(1).Info(
				"secret watch: connection has no authentication union; matching credential/TLS refs only",
				"namespace", conn.Namespace, "connection", conn.Name)
		}
		if connectionReferencesSecret(conn, secret.Name) {
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{Namespace: secret.Namespace, Name: conn.Name},
			})
		}
	}
	return reqs
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
