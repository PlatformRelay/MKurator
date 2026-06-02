package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

// ValidateTopicSpec runs admission validation for Topic spec fields.
func ValidateTopicSpec(
	ctx context.Context,
	reader client.Reader,
	namespace, resourceName string,
	spec *messagingv1alpha1.TopicSpec,
) ([]string, field.ErrorList) {
	//nolint:prealloc // error count varies by validation path
	errs := make(field.ErrorList, 0)

	errs = append(errs, ValidateKubernetesResourceName(field.NewPath("metadata").Child("name"), resourceName)...)
	errs = append(errs, ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))...)
	errs = append(errs, ValidateMQObjectName(field.NewPath("spec").Child("topicName"), spec.TopicName)...)

	warnings := unknownTopicAttributeWarnings(spec.Attributes)
	return warnings, errs
}
