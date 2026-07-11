package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// ValidateAuthorityRecordSpecV1Beta1 runs stateful admission validation for AuthorityRecord v1beta1 spec fields.
func ValidateAuthorityRecordSpecV1Beta1(
	ctx context.Context,
	reader client.Reader,
	namespace, _ string,
	spec *messagingv1beta1.AuthorityRecordSpec,
) field.ErrorList {
	return ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))
}
