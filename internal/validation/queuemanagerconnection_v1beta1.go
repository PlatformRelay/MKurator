package validation

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	// ClientCert (mTLS, AUTH-16) is a DIFFERENT stateful shape: the referenced Secret carries a
	// tls.crt/tls.key keypair, NOT username/password. Route it to a dedicated keypair check
	// (existence + both keys present) rather than the username/password path below — otherwise a
	// tls Secret would draw a spurious "missing username" warning and its keypair would never be
	// checked. Structural exclusivity of the union is CEL-first (ADR-0025), so by the time this
	// webhook runs the union is well-formed and clientCert is set iff mode is ClientCert.
	if a := spec.Authentication; a != nil &&
		a.Mode == messagingv1beta1.MQWebAuthenticationModeClientCert && a.ClientCert != nil {
		errs := validateClientCertSecret(ctx, reader, namespace, a.ClientCert.SecretRef.Name)
		return nil, errs
	}

	// Resolve the effective credentials Secret for stateful checks. Structural
	// exclusivity of the authentication union is enforced CEL-first (ADR-0025), so by the
	// time this webhook runs, the union — when present — is well-formed. We only need the
	// Secret name the connection will actually read (username/password):
	//   - authentication union, mode Basic -> authentication.basic.secretRef
	//   - authentication union, mode LTPA  -> authentication.ltpa.secretRef (AUTH-13; the
	//     LTPA login Secret carries the same username/password keys)
	//   - legacy credentialsSecretRef (union absent) -> credentialsSecretRef
	// Basic and LTPA reach here; ClientCert is handled above.
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

	alphaSpec := messagingv1beta1.QueueManagerConnectionSpec{
		QueueManager: spec.QueueManager,
		Endpoint:     spec.Endpoint,
		RESTPrefix:   spec.RESTPrefix,
		CredentialsSecretRef: &messagingv1beta1.SecretReference{
			Name: credName,
		},
	}
	if spec.TLS != nil {
		alphaSpec.TLS = &messagingv1beta1.TLSConfig{
			InsecureSkipVerify: spec.TLS.InsecureSkipVerify,
		}
		if spec.TLS.CASecretRef != nil {
			alphaSpec.TLS.CASecretRef = &messagingv1beta1.SecretReference{Name: spec.TLS.CASecretRef.Name}
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

// validateClientCertSecret performs the stateful ClientCert (mTLS) keypair check (AUTH-16,
// ADR-0027 AC2): the referenced Secret must exist and carry BOTH a non-empty tls.crt and
// tls.key. The keys-existence/shape check lives here in the STATEFUL webhook tier, never in
// CEL. Deep keypair validity (parse/match) is enforced at client-build time as a
// configuration-class error; this webhook only guards the presence/shape so a misshapen
// Secret is rejected at admission with a clear, actionable message.
func validateClientCertSecret(
	ctx context.Context,
	reader client.Reader,
	namespace, name string,
) field.ErrorList {
	path := field.NewPath("spec").Child("authentication").Child("clientCert").Child("secretRef").Child("name")
	errs, secret := getSecretOrErrors(ctx, reader, namespace, name, path)
	if len(errs) > 0 || secret == nil {
		return errs
	}
	var out field.ErrorList
	for _, key := range []string{"tls.crt", "tls.key"} {
		if v, ok := secret.Data[key]; !ok || len(v) == 0 {
			out = append(out, field.Invalid(path, name, fmt.Sprintf(
				"client certificate Secret %q must contain a non-empty %q (a kubernetes.io/tls-shaped Secret)",
				name, key)))
		}
	}
	return out
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
