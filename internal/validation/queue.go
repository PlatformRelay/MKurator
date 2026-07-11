package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// ValidateQueueSpec runs stateful admission validation for Queue spec fields.
func ValidateQueueSpec(
	ctx context.Context,
	reader client.Reader,
	namespace, _ string,
	spec *messagingv1alpha1.QueueSpec,
) ([]string, field.ErrorList) {
	errs := ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))
	warnings := unknownQueueAttributeWarnings(mqadmin.QueueType(spec.Type), spec.Attributes)
	return warnings, errs
}
