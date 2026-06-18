package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/conduit-ops/mkurator/api/v1alpha1"
	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

// ValidateChannelSpec runs stateful admission validation for Channel spec fields.
func ValidateChannelSpec(
	ctx context.Context,
	reader client.Reader,
	namespace, _ string,
	spec *messagingv1alpha1.ChannelSpec,
) ([]string, field.ErrorList) {
	errs := ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))
	errs = append(errs, validateChannelTypeRequirements(spec, field.NewPath("spec"))...)
	warnings := unknownChannelAttributeWarnings(spec.Attributes)
	return warnings, errs
}

func validateChannelTypeRequirements(
	spec *messagingv1alpha1.ChannelSpec,
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

func attrValue(attrs map[string]string, key string) string {
	if attrs == nil {
		return ""
	}
	if v, ok := attrs[key]; ok {
		return v
	}
	if v, ok := attrs[mqadmin.NormalizeAttrKey(key)]; ok {
		return v
	}
	return ""
}
