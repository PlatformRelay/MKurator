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
		name string
		// clone returns the address of a populated original and its DeepCopy(), both as *T,
		// so the loop can assert the copy deep-equals the original and is a distinct pointer.
		clone func() (orig any, cp any)
	}{
		{"SecretReference", func() (any, any) { v := SecretReference{Name: "s"}; return &v, v.DeepCopy() }},
		{"LocalObjectReference", func() (any, any) { v := ref; return &v, v.DeepCopy() }},
		{"TLSConfig", func() (any, any) {
			v := TLSConfig{InsecureSkipVerify: true, CASecretRef: &SecretReference{Name: "ca"}}
			return &v, v.DeepCopy()
		}},
		{
			"BasicAuth",
			func() (any, any) { v := BasicAuth{SecretRef: SecretReference{Name: "b"}}; return &v, v.DeepCopy() },
		},
		{
			"LTPAAuth",
			func() (any, any) { v := LTPAAuth{SecretRef: SecretReference{Name: "l"}}; return &v, v.DeepCopy() },
		},
		{
			"ClientCertAuth",
			func() (any, any) { v := ClientCertAuth{SecretRef: SecretReference{Name: "c"}}; return &v, v.DeepCopy() },
		},
		{"MQWebAuthentication", func() (any, any) {
			v := MQWebAuthentication{
				Mode:       MQWebAuthenticationModeBasic,
				Basic:      &BasicAuth{SecretRef: SecretReference{Name: "b"}},
				LTPA:       &LTPAAuth{SecretRef: SecretReference{Name: "l"}},
				ClientCert: &ClientCertAuth{SecretRef: SecretReference{Name: "c"}},
			}
			return &v, v.DeepCopy()
		}},
		{"WorkloadLifecyclePolicies", func() (any, any) { v := wlp; return &v, v.DeepCopy() }},
		{"MQObjectStatusFields", func() (any, any) { v := status; return &v, v.DeepCopy() }},
		{"QueueSpec", func() (any, any) {
			v := QueueSpec{ConnectionRef: ref, QueueName: "Q", MaxDepth: int32Ptr(10),
				Attributes: map[string]string{"k": "v"}, WorkloadLifecyclePolicies: wlp}
			return &v, v.DeepCopy()
		}},
		{"QueueStatus", func() (any, any) {
			v := QueueStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return &v, v.DeepCopy()
		}},
		{"TopicSpec", func() (any, any) {
			v := TopicSpec{
				ConnectionRef:             ref,
				TopicName:                 "T",
				Attributes:                map[string]string{"k": "v"},
				WorkloadLifecyclePolicies: wlp,
			}
			return &v, v.DeepCopy()
		}},
		{"TopicStatus", func() (any, any) {
			v := TopicStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return &v, v.DeepCopy()
		}},
		{"ChannelSpec", func() (any, any) {
			v := ChannelSpec{ConnectionRef: ref, ChannelName: "C", ShareConv: int32Ptr(10),
				MaxMsgLength: int32Ptr(4096), MaxInstances: int32Ptr(5), MaxInstancesClient: int32Ptr(2),
				Attributes: map[string]string{"k": "v"}, WorkloadLifecyclePolicies: wlp}
			return &v, v.DeepCopy()
		}},
		{"ChannelStatus", func() (any, any) {
			v := ChannelStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return &v, v.DeepCopy()
		}},
		{"ChannelAuthRuleSpec", func() (any, any) {
			v := ChannelAuthRuleSpec{ConnectionRef: ref, ChannelName: "C",
				RuleType: ChannelAuthRuleTypeAddressMap, WorkloadLifecyclePolicies: wlp}
			return &v, v.DeepCopy()
		}},
		{"ChannelAuthRuleStatus", func() (any, any) {
			v := ChannelAuthRuleStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return &v, v.DeepCopy()
		}},
		{"AuthorityRecordSpec", func() (any, any) {
			v := AuthorityRecordSpec{ConnectionRef: ref, Profile: "APP.ORDERS",
				ObjectType: AuthorityObjectTypeQueue, Principal: "app",
				Authorities: []string{"GET", "PUT"}, WorkloadLifecyclePolicies: wlp}
			return &v, v.DeepCopy()
		}},
		{"AuthorityRecordStatus", func() (any, any) {
			v := AuthorityRecordStatus{Conditions: testConditions(), MQObjectStatusFields: status}
			return &v, v.DeepCopy()
		}},
		{"QueueManagerConnectionSpec", func() (any, any) { s := fullyPopulatedQMC().Spec; return &s, s.DeepCopy() }},
		{"QueueManagerConnectionStatus", func() (any, any) {
			v := QueueManagerConnectionStatus{Conditions: testConditions(), ObservedGeneration: 3}
			return &v, v.DeepCopy()
		}},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			orig, cp := c.clone()
			if cp == nil || reflect.ValueOf(cp).IsNil() {
				t.Fatalf("%s: DeepCopy returned nil", c.name)
			}
			if orig == cp {
				t.Fatalf("%s: DeepCopy returned the same pointer as the original", c.name)
			}
			if !reflect.DeepEqual(orig, cp) {
				t.Fatalf("%s: copy not deeply equal to original", c.name)
			}
		})
	}
}
