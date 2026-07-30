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
// v1beta1 QueueManagerConnection is also in DisableFor: the reconciler and the mqweb factory
// read the QMC spec (credentials/authentication) and status directly through the manager client,
// so these reads must hit the API server and observe the latest object rather than a possibly
// stale informer snapshot. Direct reads are cheap here because v1beta1 is the storage version,
// so no conversion runs on the read path.
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
