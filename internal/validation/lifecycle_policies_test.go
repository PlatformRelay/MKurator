package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
)

func TestValidateWorkloadLifecyclePolicies_Valid(t *testing.T) {
	t.Parallel()
	errs := ValidateWorkloadLifecyclePolicies(field.NewPath("spec"), messagingv1alpha1.WorkloadLifecyclePolicies{
		DeletionPolicy: messagingv1alpha1.DeletionPolicyOrphan,
		AdoptionPolicy: messagingv1alpha1.AdoptionPolicyAdoptIfMatching,
	})
	if len(errs) != 0 {
		t.Fatalf("errs = %v", errs)
	}
}

func TestValidateWorkloadLifecyclePolicies_Invalid(t *testing.T) {
	t.Parallel()
	errs := ValidateWorkloadLifecyclePolicies(field.NewPath("spec"), messagingv1alpha1.WorkloadLifecyclePolicies{
		DeletionPolicy: "Retain", AdoptionPolicy: "Always",
	})
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %v", errs)
	}
}
