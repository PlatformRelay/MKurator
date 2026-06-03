package validation

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

// ValidateChannelAuthRuleSpec runs admission validation for ChannelAuthRule spec fields.
func ValidateChannelAuthRuleSpec(
	ctx context.Context,
	reader client.Reader,
	namespace, resourceName string,
	spec *messagingv1alpha1.ChannelAuthRuleSpec,
) field.ErrorList {
	var errs field.ErrorList

	errs = append(errs, ValidateKubernetesResourceName(field.NewPath("metadata").Child("name"), resourceName)...)
	errs = append(errs, ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))...)
	errs = append(errs, ValidateMQObjectName(field.NewPath("spec").Child("channelName"), spec.ChannelName)...)
	errs = append(errs, ValidateManagedChannelRef(ctx, reader, namespace, spec.ConnectionRef.Name, spec.ChannelName,
		field.NewPath("spec").Child("channelName"))...)

	switch spec.RuleType {
	case messagingv1alpha1.ChannelAuthRuleTypeAddressMap:
		if spec.Address == "" {
			errs = append(errs, field.Required(field.NewPath("spec").Child("address"),
				"address is required for ADDRESSMAP rules"))
		}
	case messagingv1alpha1.ChannelAuthRuleTypeBlockUser:
		if spec.UserList == "" {
			errs = append(errs, field.Required(field.NewPath("spec").Child("userList"),
				"userList is required for BLOCKUSER rules"))
		}
	case messagingv1alpha1.ChannelAuthRuleTypeBlockAddr:
		if spec.Address == "" {
			errs = append(errs, field.Required(field.NewPath("spec").Child("address"),
				"address is required for BLOCKADDR rules"))
		}
	case "":
		errs = append(errs, field.Required(field.NewPath("spec").Child("ruleType"), "ruleType is required"))
	}

	return errs
}

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
		if strings.TrimSpace(auth) == "" {
			errs = append(errs, field.Invalid(field.NewPath("spec").Child("authorities").Index(i), auth,
				"authority must not be empty"))
		}
	}

	return errs
}
