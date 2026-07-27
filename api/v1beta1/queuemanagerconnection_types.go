package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Condition types for MKurator resources.
const (
	ConditionReady  = "Ready"
	ConditionSynced = "Synced"
)

// Condition reasons shared across resources.
const (
	ReasonAvailable        = "Available"
	ReasonProgressing      = "Progressing"
	ReasonError            = "Error"
	ReasonDeleting         = "Deleting"
	ReasonDriftDetected    = "DriftDetected"
	ReasonOrphaned         = "Orphaned"
	ReasonAdoptionConflict = "AdoptionConflict"
	ReasonAlreadyExists    = "AlreadyExists"
	ReasonSuspended        = "Suspended"
)

// ReconcileRequestedAtAnnotation triggers an immediate reconcile when its value changes.
const ReconcileRequestedAtAnnotation = "messaging.mkurator.dev/reconcile-requested-at"

// ForceOrphanAnnotation skips MQ cleanup and removes the finalizer on a deleting CR.
const ForceOrphanAnnotation = "messaging.mkurator.dev/force-orphan"

// DriftPolicyAnnotation selects how the operator responds to spec vs IBM MQ drift.
const DriftPolicyAnnotation = "messaging.mkurator.dev/drift-policy"

// DriftPolicyObserveOnly reports drift without issuing DEFINE/ALTER to correct MQ.
const DriftPolicyObserveOnly = "observe-only"

// QueueManagerConnectionFinalizer ensures MQ cleanup completes before removal.
const QueueManagerConnectionFinalizer = "messaging.mkurator.dev/connection"

// AllowInsecureTLSAnnotation opts in to tls.insecureSkipVerify on QueueManagerConnection (dev only).
const AllowInsecureTLSAnnotation = "messaging.mkurator.dev/allow-insecure-tls"

// QueueFinalizer ensures the MQ queue is deleted before the CR is removed.
const QueueFinalizer = "messaging.mkurator.dev/queue"

// QueueManagerConnectionSpec defines how to reach an IBM MQ queue manager.
//
// CEL admission (ADR-0025, ADR-0027) enforces:
//   - a credential source is always present: either the legacy credentialsSecretRef
//     (implicit Basic) or an explicit authentication union.
//   - structural exclusivity of the authentication union: the member struct matching
//     mode must be set and no other member may be set. Stateful checks (Secret
//     existence and keys) stay in the validating webhook.
//
// +kubebuilder:validation:XValidation:rule="has(self.credentialsSecretRef) || has(self.authentication)",message="either credentialsSecretRef or authentication must be set"
type QueueManagerConnectionSpec struct {
	// QueueManager is the IBM MQ queue manager name (case-sensitive).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	QueueManager string `json:"queueManager"`

	// Endpoint is the mqweb base URL, e.g. https://ibm-mq.ibm-mq.svc:9443
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https://`
	Endpoint string `json:"endpoint"`

	// RESTPrefix is the mqweb REST API path prefix. Defaults to /ibmmq/rest/v3.
	// +kubebuilder:validation:Pattern=`^/`
	// +optional
	RESTPrefix string `json:"restPrefix,omitempty"`

	// TLS configures HTTPS trust for mqweb.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`

	// CredentialsSecretRef references a Secret with mqweb credentials.
	// Required key: password (mqAdminPassword is also accepted). Username is optional:
	// username, user, or mqAdminUser (defaults to admin when absent; admission warns).
	//
	// Optional as of ADR-0027: when omitted, an explicit authentication union must be
	// supplied instead (CEL-enforced). When present and authentication is absent, the
	// connection uses HTTP Basic against this Secret — the pre-ADR-0027 behaviour, so
	// every existing manifest keeps working verbatim.
	// +optional
	CredentialsSecretRef *SecretReference `json:"credentialsSecretRef,omitempty"`

	// Authentication selects how MKurator authenticates to the mqweb admin REST API
	// (ADR-0027). Exactly one mode; the member struct matching mode is required and no
	// other member may be set (CEL-enforced structural exclusivity). When omitted, the
	// connection defaults to Basic reading credentialsSecretRef, preserving existing
	// manifests. This slice ships only the Basic enum value; LTPA and ClientCert land
	// in later slices (AUTH-13+).
	// +optional
	Authentication *MQWebAuthentication `json:"authentication,omitempty"`
}

// MQWebAuthenticationMode selects the mqweb admin REST authentication mechanism.
//
// Only Basic is accepted in this slice (ADR-0027 sequencing): shipping LTPA/ClientCert
// as accepted enum values before their runtime lands would admit a spec that cannot be
// reconciled (a "dead letter"). The member structs for LTPA and ClientCert exist on the
// union so the exclusivity CEL is expressible, but selecting those modes is rejected by
// the enum until their slices ship.
// +kubebuilder:validation:Enum=Basic
type MQWebAuthenticationMode string

const (
	// MQWebAuthenticationModeBasic authenticates with HTTP Basic against a Secret (username/password).
	MQWebAuthenticationModeBasic MQWebAuthenticationMode = "Basic"
	// MQWebAuthenticationModeLTPA authenticates with an LTPA login token (AUTH-13; not yet accepted).
	MQWebAuthenticationModeLTPA MQWebAuthenticationMode = "LTPA"
	// MQWebAuthenticationModeClientCert authenticates with a client certificate/mTLS (later slice; not yet accepted).
	MQWebAuthenticationModeClientCert MQWebAuthenticationMode = "ClientCert"
)

// MQWebAuthentication is the discriminated union selecting the mqweb admin auth mode.
//
// Structural exclusivity is CEL-first (ADR-0025): the member matching mode must be set,
// and no non-matching member may be set. The validating webhook covers only stateful
// checks (Secret existence and keys), never structural exclusivity.
//
// +kubebuilder:validation:XValidation:rule="self.mode != 'Basic' || has(self.basic)",message="authentication.basic is required when mode is Basic"
// +kubebuilder:validation:XValidation:rule="self.mode != 'LTPA' || has(self.ltpa)",message="authentication.ltpa is required when mode is LTPA"
// +kubebuilder:validation:XValidation:rule="self.mode != 'ClientCert' || has(self.clientCert)",message="authentication.clientCert is required when mode is ClientCert"
// +kubebuilder:validation:XValidation:rule="!has(self.basic) || self.mode == 'Basic'",message="authentication.basic may only be set when mode is Basic"
// +kubebuilder:validation:XValidation:rule="!has(self.ltpa) || self.mode == 'LTPA'",message="authentication.ltpa may only be set when mode is LTPA"
// +kubebuilder:validation:XValidation:rule="!has(self.clientCert) || self.mode == 'ClientCert'",message="authentication.clientCert may only be set when mode is ClientCert"
type MQWebAuthentication struct {
	// Mode selects the authentication mechanism. Only Basic is accepted in this slice.
	// +kubebuilder:validation:Required
	Mode MQWebAuthenticationMode `json:"mode"`

	// Basic authenticates with HTTP Basic against SecretRef (keys: username/password,
	// mqAdminUser/mqAdminPassword also accepted). Required when mode is Basic.
	// +optional
	Basic *BasicAuth `json:"basic,omitempty"`

	// LTPA authenticates with an LTPA login token (AUTH-13). Required when mode is LTPA.
	// +optional
	LTPA *LTPAAuth `json:"ltpa,omitempty"`

	// ClientCert authenticates with a client certificate/mTLS (later slice).
	// Required when mode is ClientCert.
	// +optional
	ClientCert *ClientCertAuth `json:"clientCert,omitempty"`
}

// BasicAuth configures HTTP Basic authentication to the mqweb admin REST API.
type BasicAuth struct {
	// SecretRef references a Secret with mqweb credentials.
	// Required key: password (mqAdminPassword is also accepted). Username is optional:
	// username, user, or mqAdminUser (defaults to admin when absent; admission warns).
	// +kubebuilder:validation:Required
	SecretRef SecretReference `json:"secretRef"`
}

// LTPAAuth configures LTPA login-token authentication (AUTH-13; runtime not yet wired).
type LTPAAuth struct {
	// SecretRef references a Secret holding the login username/password used to obtain
	// an LTPA cookie. LTPA is login-derived (ADR-0027).
	// +kubebuilder:validation:Required
	SecretRef SecretReference `json:"secretRef"`
}

// ClientCertAuth configures client-certificate/mTLS authentication (later slice; runtime not yet wired).
type ClientCertAuth struct {
	// SecretRef references a Secret holding the client keypair (keys tls.crt/tls.key).
	// +kubebuilder:validation:Required
	SecretRef SecretReference `json:"secretRef"`
}

// TLSConfig holds TLS options for mqweb.
type TLSConfig struct {
	// InsecureSkipVerify disables server certificate verification (dev only).
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// CASecretRef references a Secret containing a CA bundle (key tls.crt, ca.crt, or ca.pem).
	// +optional
	CASecretRef *SecretReference `json:"caSecretRef,omitempty"`
}

// SecretReference identifies a Secret in the same namespace as the CR.
type SecretReference struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// QueueManagerConnectionStatus defines the observed state of QueueManagerConnection.
type QueueManagerConnectionStatus struct {
	// Conditions represent the current state of the connection.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation last reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=qmc
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// QueueManagerConnection describes how to reach an IBM MQ queue manager via mqweb.
type QueueManagerConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QueueManagerConnectionSpec   `json:"spec,omitempty"`
	Status QueueManagerConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QueueManagerConnectionList contains a list of QueueManagerConnection.
type QueueManagerConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QueueManagerConnection `json:"items"`
}

func init() {
	Register(&QueueManagerConnection{}, &QueueManagerConnectionList{})
}
