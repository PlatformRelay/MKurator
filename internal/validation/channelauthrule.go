package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
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
	errs = append(errs, ValidateWorkloadLifecyclePolicies(field.NewPath("spec"), spec.WorkloadLifecyclePolicies)...)
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

	errs = append(errs, ValidateChannelAuthUserSource(field.NewPath("spec").Child("userSource"), spec.UserSource)...)
	errs = append(errs, ValidateChannelAuthCheckClient(field.NewPath("spec").Child("checkClient"), spec.CheckClient)...)

	return errs
}
