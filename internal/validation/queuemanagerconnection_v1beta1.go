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
	// Resolve the effective credentials Secret for stateful checks. Structural
	// exclusivity of the authentication union is enforced CEL-first (ADR-0025), so by the
	// time this webhook runs, the union — when present — is well-formed. We only need the
	// Secret name the connection will actually read (username/password):
	//   - authentication union, mode Basic -> authentication.basic.secretRef
	//   - authentication union, mode LTPA  -> authentication.ltpa.secretRef (AUTH-13; the
	//     LTPA login Secret carries the same username/password keys)
	//   - legacy credentialsSecretRef (union absent) -> credentialsSecretRef
	// Basic and LTPA are the accepted enum values, so no other mode reaches here.
	credName := ""
	if spec.CredentialsSecretRef != nil {
		credName = spec.CredentialsSecretRef.Name
	}
	if a := spec.Authentication; a != nil {
		switch {
		case a.Mode == messagingv1beta1.MQWebAuthenticationModeBasic && a.Basic != nil:
			credName = a.Basic.SecretRef.Name
		case a.Mode == messagingv1beta1.MQWebAuthenticationModeLTPA && a.LTPA != nil:
			credName = a.LTPA.SecretRef.Name
		}
	}

	alphaSpec := messagingv1alpha1.QueueManagerConnectionSpec{
		QueueManager: spec.QueueManager,
		Endpoint:     spec.Endpoint,
		RESTPrefix:   spec.RESTPrefix,
		CredentialsSecretRef: messagingv1alpha1.SecretReference{
			Name: credName,
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
