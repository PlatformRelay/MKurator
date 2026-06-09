package validation

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
)

// ValidateAuthorityRecordSpec runs admission validation for AuthorityRecord spec fields.
func ValidateAuthorityRecordSpec(
	ctx context.Context,
	reader client.Reader,
	namespace, resourceName string,
	spec *messagingv1alpha1.AuthorityRecordSpec,
) field.ErrorList {
	var errs field.ErrorList

	errs = append(errs, ValidateKubernetesResourceName(field.NewPath("metadata").Child("name"), resourceName)...)
	errs = append(errs, ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))...)
	errs = append(errs, ValidateWorkloadLifecyclePolicies(field.NewPath("spec"), spec.WorkloadLifecyclePolicies)...)
	errs = append(errs, ValidateMQObjectName(field.NewPath("spec").Child("profile"), spec.Profile)...)

	if spec.ObjectType == "" {
		errs = append(errs, field.Required(field.NewPath("spec").Child("objectType"), "objectType is required"))
	}
	if spec.Principal == "" && spec.Group == "" {
		errs = append(errs, field.Required(field.NewPath("spec").Child("principal"),
			"principal or group is required"))
	}
	if spec.Principal != "" && spec.Group != "" {
		errs = append(errs, field.Invalid(field.NewPath("spec").Child("group"), spec.Group,
			"specify principal or group, not both"))
	}
	if len(spec.Authorities) == 0 {
		errs = append(errs, field.Required(field.NewPath("spec").Child("authorities"),
			"at least one authority is required"))
	}
	for i, auth := range spec.Authorities {
		path := field.NewPath("spec").Child("authorities").Index(i)
		if strings.TrimSpace(auth) == "" {
			errs = append(errs, field.Invalid(path, auth, "authority must not be empty"))
			continue
		}
		errs = append(errs, ValidateMQAuthorityKeyword(path, auth)...)
	}

	return errs
}
