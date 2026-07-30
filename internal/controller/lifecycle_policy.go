package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// workloadLifecyclePolicies returns the object's lifecycle policies in the v1alpha1-typed shape
// the shared drift/adoption helpers compare against. The migrated v1beta1 Queue (8e-3) is adapted
// here via value-identical string casts so the not-yet-migrated (8e-4..7) v1alpha1 cases are
// untouched; the v1alpha1 typing collapses when v1alpha1 is dropped (8e-8).
func workloadLifecyclePolicies(obj client.Object) messagingv1alpha1.WorkloadLifecyclePolicies {
	switch o := obj.(type) {
	case *messagingv1beta1.Queue:
		return messagingv1alpha1.WorkloadLifecyclePolicies{
			DeletionPolicy: messagingv1alpha1.DeletionPolicy(o.Spec.DeletionPolicy),
			AdoptionPolicy: messagingv1alpha1.AdoptionPolicy(o.Spec.AdoptionPolicy),
		}
	case *messagingv1alpha1.Queue:
		return o.Spec.WorkloadLifecyclePolicies
	case *messagingv1beta1.Topic:
		return messagingv1alpha1.WorkloadLifecyclePolicies{
			DeletionPolicy: messagingv1alpha1.DeletionPolicy(o.Spec.DeletionPolicy),
			AdoptionPolicy: messagingv1alpha1.AdoptionPolicy(o.Spec.AdoptionPolicy),
		}
	case *messagingv1alpha1.Topic:
		return o.Spec.WorkloadLifecyclePolicies
	case *messagingv1beta1.Channel:
		return messagingv1alpha1.WorkloadLifecyclePolicies{
			DeletionPolicy: messagingv1alpha1.DeletionPolicy(o.Spec.DeletionPolicy),
			AdoptionPolicy: messagingv1alpha1.AdoptionPolicy(o.Spec.AdoptionPolicy),
		}
	case *messagingv1alpha1.Channel:
		return o.Spec.WorkloadLifecyclePolicies
	case *messagingv1alpha1.ChannelAuthRule:
		return o.Spec.WorkloadLifecyclePolicies
	case *messagingv1alpha1.AuthorityRecord:
		return o.Spec.WorkloadLifecyclePolicies
	default:
		return messagingv1alpha1.WorkloadLifecyclePolicies{}
	}
}

func workloadDeletionPolicy(obj client.Object) messagingv1alpha1.DeletionPolicy {
	policy := workloadLifecyclePolicies(obj).DeletionPolicy
	if policy == "" {
		return messagingv1alpha1.DeletionPolicyDelete
	}
	return policy
}

func workloadAdoptionPolicy(obj client.Object) messagingv1alpha1.AdoptionPolicy {
	policy := workloadLifecyclePolicies(obj).AdoptionPolicy
	if policy == "" {
		return messagingv1alpha1.AdoptionPolicyAdopt
	}
	return policy
}

func workloadFirstAdoption(obj client.Object) bool {
	switch o := obj.(type) {
	case *messagingv1beta1.Queue:
		return o.Status.ObservedGeneration == 0
	case *messagingv1alpha1.Queue:
		return o.Status.ObservedGeneration == 0
	case *messagingv1beta1.Topic:
		return o.Status.ObservedGeneration == 0
	case *messagingv1alpha1.Topic:
		return o.Status.ObservedGeneration == 0
	case *messagingv1beta1.Channel:
		return o.Status.ObservedGeneration == 0
	case *messagingv1alpha1.Channel:
		return o.Status.ObservedGeneration == 0
	case *messagingv1alpha1.ChannelAuthRule:
		return o.Status.ObservedGeneration == 0
	case *messagingv1alpha1.AuthorityRecord:
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
	return workloadDeletionPolicy(co) == messagingv1alpha1.DeletionPolicyOrphan
}
