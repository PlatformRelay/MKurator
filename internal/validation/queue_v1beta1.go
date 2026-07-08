package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

// ValidateQueueSpecV1Beta1 runs stateful admission validation for Queue v1beta1 spec fields.
func ValidateQueueSpecV1Beta1(
	ctx context.Context,
	reader client.Reader,
	namespace, _ string,
	spec *messagingv1beta1.QueueSpec,
) ([]string, field.ErrorList) {
	errs := ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))
	warnings := deprecatedQueueAttributeWarnings(spec.Attributes)
	warnings = append(warnings, unknownQueueAttributeWarnings(mqadmin.QueueType(spec.Type), spec.Attributes)...)
	return warnings, errs
}
