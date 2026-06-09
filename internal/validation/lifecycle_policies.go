package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
)

func ValidateWorkloadLifecyclePolicies(path *field.Path, policies messagingv1alpha1.WorkloadLifecyclePolicies) field.ErrorList {
	var errs field.ErrorList
	if policies.DeletionPolicy != "" &&
		policies.DeletionPolicy != messagingv1alpha1.DeletionPolicyDelete &&
		policies.DeletionPolicy != messagingv1alpha1.DeletionPolicyOrphan {
		errs = append(errs, field.Invalid(path.Child("deletionPolicy"), policies.DeletionPolicy,
			"must be Delete or Orphan"))
	}
	if policies.AdoptionPolicy != "" &&
		policies.AdoptionPolicy != messagingv1alpha1.AdoptionPolicyAdopt &&
		policies.AdoptionPolicy != messagingv1alpha1.AdoptionPolicyAdoptIfMatching &&
		policies.AdoptionPolicy != messagingv1alpha1.AdoptionPolicyFailIfExists {
		errs = append(errs, field.Invalid(path.Child("adoptionPolicy"), policies.AdoptionPolicy,
			"must be Adopt, AdoptIfMatching, or FailIfExists"))
	}
	return errs
}
