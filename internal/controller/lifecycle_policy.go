package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// workloadLifecyclePolicies returns the object's lifecycle policies from the v1beta1 spec,
// which embeds WorkloadLifecyclePolicies inline. The shared drift/adoption helpers compare
// against this shape.
func workloadLifecyclePolicies(obj client.Object) messagingv1beta1.WorkloadLifecyclePolicies {
	switch o := obj.(type) {
	case *messagingv1beta1.Queue:
		return o.Spec.WorkloadLifecyclePolicies
	case *messagingv1beta1.Topic:
		return o.Spec.WorkloadLifecyclePolicies
	case *messagingv1beta1.Channel:
		return o.Spec.WorkloadLifecyclePolicies
	case *messagingv1beta1.ChannelAuthRule:
		return o.Spec.WorkloadLifecyclePolicies
	case *messagingv1beta1.AuthorityRecord:
		return o.Spec.WorkloadLifecyclePolicies
	default:
		return messagingv1beta1.WorkloadLifecyclePolicies{}
	}
}

func workloadDeletionPolicy(obj client.Object) messagingv1beta1.DeletionPolicy {
	policy := workloadLifecyclePolicies(obj).DeletionPolicy
	if policy == "" {
		return messagingv1beta1.DeletionPolicyDelete
	}
	return policy
}

func workloadAdoptionPolicy(obj client.Object) messagingv1beta1.AdoptionPolicy {
	policy := workloadLifecyclePolicies(obj).AdoptionPolicy
	if policy == "" {
		return messagingv1beta1.AdoptionPolicyAdopt
	}
	return policy
}

func workloadFirstAdoption(obj client.Object) bool {
	switch o := obj.(type) {
	case *messagingv1beta1.Queue:
		return o.Status.ObservedGeneration == 0
	case *messagingv1beta1.Topic:
		return o.Status.ObservedGeneration == 0
	case *messagingv1beta1.Channel:
		return o.Status.ObservedGeneration == 0
	case *messagingv1beta1.ChannelAuthRule:
		return o.Status.ObservedGeneration == 0
	case *messagingv1beta1.AuthorityRecord:
		return o.Status.ObservedGeneration == 0
	default:
		return false
	}
}

func orphanDeletionRequested(obj metav1.Object) bool {
	if forceOrphanRequested(obj) {
		return true
	}
	co, ok := obj.(client.Object)
	if !ok {
		return false
	}
	return workloadDeletionPolicy(co) == messagingv1beta1.DeletionPolicyOrphan
}
