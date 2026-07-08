package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

// ValidateChannelAuthRuleSpecV1Beta1 runs stateful admission validation for ChannelAuthRule v1beta1 spec fields.
func ValidateChannelAuthRuleSpecV1Beta1(
	ctx context.Context,
	reader client.Reader,
	namespace, _ string,
	spec *messagingv1beta1.ChannelAuthRuleSpec,
) field.ErrorList {
	errs := ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))
	// BLOCKADDR listener rules use channelName '*' on MQ; no managed Channel CR applies.
	if spec.RuleType == messagingv1beta1.ChannelAuthRuleTypeBlockAddr && spec.ChannelName == "*" {
		return errs
	}
	return append(errs,
		ValidateManagedChannelRef(ctx, reader, namespace, spec.ConnectionRef.Name, spec.ChannelName,
			field.NewPath("spec").Child("channelName"))...,
	)
}
