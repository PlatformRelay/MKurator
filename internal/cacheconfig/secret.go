package cacheconfig

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
// ARCHITECTURE.md least-privilege narrative update is deferred to ROADMAP Phase 7d (Wave 4).
func ManagerOptions() (cache.Options, client.Options) {
	return cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&corev1.Secret{}: {
					Transform: stripSecretCredentialTransform,
				},
			},
		}, client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{&corev1.Secret{}},
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
