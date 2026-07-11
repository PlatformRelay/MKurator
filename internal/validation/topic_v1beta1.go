package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// ValidateTopicSpecV1Beta1 runs stateful admission validation for Topic v1beta1 spec fields.
func ValidateTopicSpecV1Beta1(
	ctx context.Context,
	reader client.Reader,
	namespace, _ string,
	spec *messagingv1beta1.TopicSpec,
) ([]string, field.ErrorList) {
	errs := ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))
	warnings := deprecatedTopicAttributeWarnings(spec.Attributes)
	warnings = append(warnings, unknownTopicAttributeWarnings(spec.Attributes)...)
	return warnings, errs
}
