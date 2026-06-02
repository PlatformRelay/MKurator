package validation

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

// ValidateQueueManagerConnectionSpec runs admission validation for QueueManagerConnection spec fields.
func ValidateQueueManagerConnectionSpec(
	ctx context.Context,
	reader client.Reader,
	namespace string,
	spec *messagingv1alpha1.QueueManagerConnectionSpec,
) field.ErrorList {
	var errs field.ErrorList

	if spec.QueueManager == "" {
		errs = append(errs, field.Required(field.NewPath("spec").Child("queueManager"), "queueManager is required"))
	}
	if spec.Endpoint == "" {
		errs = append(errs, field.Required(field.NewPath("spec").Child("endpoint"), "endpoint is required"))
	} else if !strings.HasPrefix(spec.Endpoint, "https://") {
		errs = append(errs, field.Invalid(field.NewPath("spec").Child("endpoint"), spec.Endpoint,
			"endpoint must use HTTPS (https://)"))
	}
	if spec.CredentialsSecretRef.Name == "" {
		errs = append(errs, field.Required(field.NewPath("spec").Child("credentialsSecretRef").Child("name"),
			"credentialsSecretRef.name is required"))
	} else {
		errs = append(errs, validateSecretExists(ctx, reader, namespace,
			spec.CredentialsSecretRef.Name,
			field.NewPath("spec").Child("credentialsSecretRef").Child("name"))...)
	}

	if spec.TLS != nil && spec.TLS.CASecretRef != nil && spec.TLS.CASecretRef.Name != "" {
		errs = append(errs, validateSecretExists(ctx, reader, namespace,
			spec.TLS.CASecretRef.Name,
			field.NewPath("spec").Child("tls").Child("caSecretRef").Child("name"))...)
	}

	return errs
}

func validateSecretExists(
	ctx context.Context,
	reader client.Reader,
	namespace, name string,
	path *field.Path,
) field.ErrorList {
	secret := &corev1.Secret{}
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := reader.Get(ctx, key, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return field.ErrorList{
				field.NotFound(path, fmt.Sprintf("Secret %q not found in namespace %q", name, namespace)),
			}
		}
		return field.ErrorList{field.InternalError(path, fmt.Errorf("get Secret %q: %w", name, err))}
	}
	return nil
}
