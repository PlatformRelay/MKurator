package validation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
	"github.com/konih/mkurator/internal/mqadmin"
)

// ValidateQueueSpec runs admission validation for Queue spec fields.
func ValidateQueueSpec(
	ctx context.Context,
	reader client.Reader,
	namespace, resourceName string,
	spec *messagingv1alpha1.QueueSpec,
) ([]string, field.ErrorList) {
	var (
		//nolint:prealloc // warnings count depends on attribute keys
		warnings = make([]string, 0)
		errs     field.ErrorList
	)

	errs = append(errs, ValidateKubernetesResourceName(field.NewPath("metadata").Child("name"), resourceName)...)
	errs = append(errs, ValidateConnectionRef(ctx, reader, namespace, spec.ConnectionRef.Name,
		field.NewPath("spec").Child("connectionRef").Child("name"))...)
	errs = append(errs, ValidateMQObjectName(field.NewPath("spec").Child("queueName"), spec.QueueName)...)

	normalized := normalizeQueueAttributes(spec.Attributes, spec.Type)
	switch mqadmin.NormalizeQueueType(mqadmin.QueueType(spec.Type)) {
	case mqadmin.QueueTypeAlias:
		if v := normalized["targq"]; v == "" {
			errs = append(errs, field.Required(field.NewPath("spec").Child("attributes").Key("targq"),
				"alias queues require attribute targq (or target)"))
		}
	case mqadmin.QueueTypeRemote:
		if v := normalized["xmitq"]; v == "" {
			errs = append(errs, field.Required(field.NewPath("spec").Child("attributes").Key("xmitq"),
				"remote queues require attribute xmitq (or transmissionqueue)"))
		}
		if v := normalized["rqmname"]; v == "" {
			errs = append(errs, field.Required(field.NewPath("spec").Child("attributes").Key("rqmname"),
				"remote queues require attribute rqmname (or remotemanager)"))
		}
	}

	warnings = append(warnings, unknownQueueAttributeWarnings(spec.Type, spec.Attributes)...)
	return warnings, errs
}

func normalizeQueueAttributes(attrs map[string]string, qType messagingv1alpha1.QueueType) map[string]string {
	normalized := make(map[string]string, len(attrs))
	for k, v := range attrs {
		normalized[mqadmin.NormalizeAttrKey(k)] = v
	}
	switch mqadmin.NormalizeQueueType(mqadmin.QueueType(qType)) {
	case mqadmin.QueueTypeAlias:
		if v, ok := normalized["target"]; ok && normalized["targq"] == "" {
			normalized["targq"] = v
		}
	case mqadmin.QueueTypeRemote:
		if v, ok := normalized["remotequeue"]; ok && normalized["rname"] == "" {
			normalized["rname"] = v
		}
		if v, ok := normalized["remotemanager"]; ok && normalized["rqmname"] == "" {
			normalized["rqmname"] = v
		}
		if v, ok := normalized["transmissionqueue"]; ok && normalized["xmitq"] == "" {
			normalized["xmitq"] = v
		}
	default:
	}
	return normalized
}
