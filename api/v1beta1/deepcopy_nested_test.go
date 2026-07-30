package v1beta1

import (
	"reflect"
	"testing"
)

// fullyPopulatedQMC returns a QueueManagerConnection with every optional pointer member set
// (TLS + CASecretRef, CredentialsSecretRef, and all three authentication members). Structural
// exclusivity of the authentication union is a CEL/webhook concern, not a deepcopy concern, so
// populating every member here is legal and exercises each generated DeepCopyInto nil-branch.
func fullyPopulatedQMC() *QueueManagerConnection {
	return &QueueManagerConnection{
		ObjectMeta: testMeta(),
		Spec: QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://mq.example:9443",
			RESTPrefix:   "/ibmmq/rest/v3",
			TLS: &TLSConfig{
				InsecureSkipVerify: true,
				CASecretRef:        &SecretReference{Name: "ca"},
			},
			CredentialsSecretRef: &SecretReference{Name: "creds"},
			Authentication: &MQWebAuthentication{
				Mode:       MQWebAuthenticationModeBasic,
				Basic:      &BasicAuth{SecretRef: SecretReference{Name: "basic"}},
				LTPA:       &LTPAAuth{SecretRef: SecretReference{Name: "ltpa"}},
				ClientCert: &ClientCertAuth{SecretRef: SecretReference{Name: "cert"}},
			},
		},
		Status: QueueManagerConnectionStatus{
			Conditions:         testConditions(),
			ObservedGeneration: 3,
		},
	}
}

// TestDeepCopyObjectFullyPopulatedQMC round-trips a QMC whose every optional pointer member is
// set, so the generated DeepCopyInto copies (rather than skips) each nested pointer. Mutating a
// deeply-nested field on the original must not leak into the clone.
func TestDeepCopyObjectFullyPopulatedQMC(t *testing.T) {
	t.Parallel()

	orig := fullyPopulatedQMC()
	clone, ok := orig.DeepCopyObject().(*QueueManagerConnection)
	if !ok {
		t.Fatal("DeepCopyObject did not return *QueueManagerConnection")
	}
	if clone == orig {
		t.Fatal("DeepCopyObject returned the same pointer")
	}
	if !reflect.DeepEqual(orig, clone) {
		t.Fatal("clone not deeply equal to original")
	}

	// Mutating a deeply-nested pointer field on the original must not affect the clone.
	orig.Spec.TLS.CASecretRef.Name = "mutated"
	if clone.Spec.TLS.CASecretRef.Name == "mutated" {
		t.Fatal("mutation of original TLS.CASecretRef leaked into clone")
	}
	orig.Spec.Authentication.Basic.SecretRef.Name = "mutated"
	if clone.Spec.Authentication.Basic.SecretRef.Name == "mutated" {
		t.Fatal("mutation of original Authentication.Basic leaked into clone")
	}
}

// TestValueTypeDeepCopy exercises the value-receiver DeepCopy() method on every standalone
// (non-root) generated type. The object round-trip only reaches these via DeepCopyInto, never
// through their own DeepCopy(), so they are covered here directly.
func TestValueTypeDeepCopy(t *testing.T) {
	t.Parallel()

	ref := LocalObjectReference{Name: "qm1"}
	wlp := WorkloadLifecyclePolicies{DeletionPolicy: DeletionPolicyOrphan, AdoptionPolicy: AdoptionPolicyAdopt}
	status := testStatusFields()

	checks := []struct {
		name  string
		clone func() any
	}{
		{"SecretReference", func() any { v := SecretReference{Name: "s"}; return v.DeepCopy() }},
		{"LocalObjectReference", func() any { v := ref; return v.DeepCopy() }},
		{"TLSConfig", func() any {
			v := TLSConfig{InsecureSkipVerify: true, CASecretRef: &SecretReference{Name: "ca"}}
			return v.DeepCopy()
		}},
		{"BasicAuth", func() any { v := BasicAuth{SecretRef: SecretReference{Name: "b"}}; return v.DeepCopy() }},
		{"LTPAAuth", func() any { v := LTPAAuth{SecretRef: SecretReference{Name: "l"}}; return v.DeepCopy() }},
		{
			"ClientCertAuth",
			func() any { v := ClientCertAuth{SecretRef: SecretReference{Name: "c"}}; return v.DeepCopy() },
		},
		{"MQWebAuthentication", func() any {
			v := MQWebAuthentication{
				Mode:       MQWebAuthenticationModeBasic,
				Basic:      &BasicAuth{SecretRef: SecretReference{Name: "b"}},
				LTPA:       &LTPAAuth{SecretRef: SecretReference{Name: "l"}},
				ClientCert: &ClientCertAuth{SecretRef: SecretReference{Name: "c"}},
			}
			return v.DeepCopy()
		}},
		{"WorkloadLifecyclePolicies", func() any { v := wlp; return v.DeepCopy() }},
		{"MQObjectStatusFields", func() any { v := status; return v.DeepCopy() }},
		{"QueueSpec", func() any {
			v := QueueSpec{ConnectionRef: ref, QueueName: "Q", MaxDepth: int32Ptr(10),
				Attributes: map[string]string{"k": "v"}, WorkloadLifecyclePolicies: wlp}
			return v.DeepCopy()
		}},
		{"QueueStatus", func() any {
			v := QueueStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return v.DeepCopy()
		}},
		{"TopicSpec", func() any {
			v := TopicSpec{
				ConnectionRef:             ref,
				TopicName:                 "T",
				Attributes:                map[string]string{"k": "v"},
				WorkloadLifecyclePolicies: wlp,
			}
			return v.DeepCopy()
		}},
		{"TopicStatus", func() any {
			v := TopicStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return v.DeepCopy()
		}},
		{"ChannelSpec", func() any {
			v := ChannelSpec{ConnectionRef: ref, ChannelName: "C", ShareConv: int32Ptr(10),
				MaxMsgLength: int32Ptr(4096), MaxInstances: int32Ptr(5), MaxInstancesClient: int32Ptr(2),
				Attributes: map[string]string{"k": "v"}, WorkloadLifecyclePolicies: wlp}
			return v.DeepCopy()
		}},
		{"ChannelStatus", func() any {
			v := ChannelStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return v.DeepCopy()
		}},
		{"ChannelAuthRuleSpec", func() any {
			v := ChannelAuthRuleSpec{ConnectionRef: ref, ChannelName: "C",
				RuleType: ChannelAuthRuleTypeAddressMap, WorkloadLifecyclePolicies: wlp}
			return v.DeepCopy()
		}},
		{"ChannelAuthRuleStatus", func() any {
			v := ChannelAuthRuleStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return v.DeepCopy()
		}},
		{"AuthorityRecordSpec", func() any {
			v := AuthorityRecordSpec{ConnectionRef: ref, Profile: "APP.ORDERS",
				ObjectType: AuthorityObjectTypeQueue, Principal: "app",
				Authorities: []string{"GET", "PUT"}, WorkloadLifecyclePolicies: wlp}
			return v.DeepCopy()
		}},
		{"AuthorityRecordStatus", func() any {
			v := AuthorityRecordStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return v.DeepCopy()
		}},
		{"QueueManagerConnectionSpec", func() any { return fullyPopulatedQMC().Spec.DeepCopy() }},
		{"QueueManagerConnectionStatus", func() any {
			v := QueueManagerConnectionStatus{Conditions: testConditions(), ObservedGeneration: 3}
			return v.DeepCopy()
		}},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := c.clone()
			if got == nil || reflect.ValueOf(got).IsNil() {
				t.Fatalf("%s: DeepCopy returned nil", c.name)
			}
		})
	}
}
