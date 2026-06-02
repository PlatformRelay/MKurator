package validation

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

// ValidateChannelSpec runs admission validation for Channel spec fields.
func ValidateChannelSpec(
	ctx context.Context,
	reader client.Reader,
	namespace, resourceName string,
	spec *messagingv1alpha1.ChannelSpec,
) ([]string, field.ErrorList) {
	var errs field.ErrorList

	errs = append(errs, ValidateKubernetesResourceName(field.NewPath("metadata").Child("name"), resourceName)...)
	errs = append(errs, ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))...)
	errs = append(errs, ValidateMQObjectName(field.NewPath("spec").Child("channelName"), spec.ChannelName)...)

	if spec.Type != "" && spec.Type != messagingv1alpha1.ChannelTypeSvrconn {
		errs = append(errs, field.Invalid(field.NewPath("spec").Child("type"), spec.Type,
			fmt.Sprintf("channel type %q is not supported in v1alpha1", spec.Type)))
	}

	warnings := unknownChannelAttributeWarnings(spec.Attributes)
	return warnings, errs
}
