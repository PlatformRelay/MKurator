package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

// ValidateChannelSpecV1Beta1 runs stateful admission validation for Channel v1beta1 spec fields.
func ValidateChannelSpecV1Beta1(
	ctx context.Context,
	reader client.Reader,
	namespace, _ string,
	spec *messagingv1beta1.ChannelSpec,
) ([]string, field.ErrorList) {
	errs := ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))
	errs = append(errs, validateChannelTypeRequirementsV1Beta1(spec, field.NewPath("spec"))...)
	warnings := deprecatedChannelAttributeWarnings(spec.Attributes)
	warnings = append(warnings, unknownChannelAttributeWarnings(spec.Attributes)...)
	return warnings, errs
}

func validateChannelTypeRequirementsV1Beta1(
	spec *messagingv1beta1.ChannelSpec,
	path *field.Path,
) field.ErrorList {
	chType := mqadmin.NormalizeChannelType(mqadmin.ChannelType(spec.Type))
	var errs field.ErrorList
	switch chType {
	case mqadmin.ChannelTypeSdr:
		if spec.ConnName == "" && attrValue(spec.Attributes, "conname") == "" {
			errs = append(errs, field.Required(path.Child("connName"),
				"SDR channels require connName or attributes.conname"))
		}
		if spec.XmitQueue == "" && attrValue(spec.Attributes, "xmitq") == "" {
			errs = append(errs, field.Required(path.Child("xmitQueue"),
				"SDR channels require xmitQueue or attributes.xmitq"))
		}
	default:
	}
	return errs
}
