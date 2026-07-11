package validation

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// ValidateQueueManagerConnectionSpecV1Beta1 runs stateful admission validation for QueueManagerConnection v1beta1.
func ValidateQueueManagerConnectionSpecV1Beta1(
	ctx context.Context,
	reader client.Reader,
	namespace string,
	annotations map[string]string,
	spec *messagingv1beta1.QueueManagerConnectionSpec,
) ([]string, field.ErrorList) {
	alphaSpec := messagingv1alpha1.QueueManagerConnectionSpec{
		QueueManager: spec.QueueManager,
		Endpoint:     spec.Endpoint,
		RESTPrefix:   spec.RESTPrefix,
		CredentialsSecretRef: messagingv1alpha1.SecretReference{
			Name: spec.CredentialsSecretRef.Name,
		},
	}
	if spec.TLS != nil {
		alphaSpec.TLS = &messagingv1alpha1.TLSConfig{
			InsecureSkipVerify: spec.TLS.InsecureSkipVerify,
		}
		if spec.TLS.CASecretRef != nil {
			alphaSpec.TLS.CASecretRef = &messagingv1alpha1.SecretReference{Name: spec.TLS.CASecretRef.Name}
		}
	}

	return ValidateQueueManagerConnectionSpec(
		ctx,
		reader,
		namespace,
		annotations,
		&alphaSpec,
	)
}

// ValidateQueueManagerConnectionDeleteV1Beta1 denies delete while dependents still reference this connection.
func ValidateQueueManagerConnectionDeleteV1Beta1(
	ctx context.Context,
	reader client.Reader,
	conn *messagingv1beta1.QueueManagerConnection,
) field.ErrorList {
	path := field.NewPath("metadata").Child("name")
	dependents, errs := listConnectionDependents(ctx, reader, conn.Namespace, conn.Name)
	if len(errs) > 0 {
		return errs
	}
	if len(dependents) == 0 {
		return nil
	}
	return field.ErrorList{
		field.Invalid(path, conn.Name, fmt.Sprintf(
			"cannot delete QueueManagerConnection %q: %s; delete or re-point dependents first",
			conn.Name, formatDependents(dependents),
		)),
	}
}
