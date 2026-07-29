package cacheconfig

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// ManagerOptions returns cache and client options that scope Secret informer retention
// per ADR-0023 §Secret scoping (ARCH-05).
//
// RBAC: the generated ClusterRole still grants get/list/watch on Secrets cluster-wide
// (+kubebuilder:rbac on QueueManagerConnection reconciler) so admission can verify
// Secret refs exist and the QMC Secret watch receives rotation events. The transform
// strips .Data/.StringData before objects enter the informer store so the operator
// does not retain credential bytes for unrelated Secrets in memory. Secret reads through
// the manager client bypass the cache (DisableFor) and always hit the API server.
//
// v1beta1 QueueManagerConnection is also in DisableFor: resolveAuthentication re-reads the
// hub to recover the authentication union (AUTH-12/13/16) which the v1alpha1 spoke drops on
// down-conversion. The on-demand v1beta1 informer can race the first reconcile and briefly
// miss the object (IsNotFound); direct API-server reads close that window. Direct reads are
// safe here because v1beta1 is the storage version (no conversion overhead) and the re-read
// happens once per reconcile cycle, not on every method call.
//
// This does NOT cover the AUTH-14 post-merge regression where a union QMC's authentication
// was empty even on a direct hub read: that was the object's stored spec genuinely losing
// the union, via the v1alpha1 spoke's ConvertTo nil-ing it on a metadata-only round trip
// (e.g. the reconciler's finalizer-add Update). See
// authenticationUnionSnapshotAnnotation in api/v1alpha1/conversion_auth.go for the fix — no
// cache option can paper over the stored object itself being wrong.
//
// See docs/ARCHITECTURE.md "RBAC & least privilege" for the least-privilege narrative.
func ManagerOptions() (cache.Options, client.Options) {
	return cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&corev1.Secret{}: {
					Transform: stripSecretCredentialTransform,
				},
			},
		}, client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&corev1.Secret{},
					&messagingv1beta1.QueueManagerConnection{},
				},
			},
		}
}

func stripSecretCredentialTransform(obj any) (any, error) {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return obj, nil
	}
	stripped := secret.DeepCopy()
	stripped.Data = nil
	stripped.StringData = nil
	return stripped, nil
}
