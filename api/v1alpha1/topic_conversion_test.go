package v1alpha1

import (
	"testing"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

func TestTopicConvertToFromRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alpha      *Topic
		assertBeta func(t *testing.T, beta *messagingv1beta1.Topic)
		assertBack func(t *testing.T, orig, back *Topic)
	}{
		{
			name: "folds topstr attribute to typed topicString",
			alpha: &Topic{
				ObjectMeta: testObjectMeta("t-fold"),
				Spec: TopicSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					TopicName:     "RETAIL.ORDERS",
					Attributes: map[string]string{
						"topstr": "retail/orders",
						"extra":  "x",
					},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Topic) {
				t.Helper()
				if beta.Spec.TopicString != "retail/orders" {
					t.Fatalf("topicString = %q", beta.Spec.TopicString)
				}
				if mapHasKey(beta.Spec.Attributes, "topstr") {
					t.Fatalf("promoted key topstr should be removed: %v", beta.Spec.Attributes)
				}
				if beta.Spec.Attributes["extra"] != "x" {
					t.Fatalf("attributes = %v", beta.Spec.Attributes)
				}
			},
			assertBack: func(t *testing.T, orig, back *Topic) {
				t.Helper()
				if back.Spec.TopicString != "retail/orders" {
					t.Fatalf("topicString = %q", back.Spec.TopicString)
				}
				if back.Spec.Attributes["extra"] != "x" {
					t.Fatalf("attributes = %v", back.Spec.Attributes)
				}
			},
		},
		{
			name: "folds topicstr alias key",
			alpha: &Topic{
				ObjectMeta: testObjectMeta("t-topicstr"),
				Spec: TopicSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					TopicName:     "RETAIL.ORDERS",
					Attributes:    map[string]string{"topicstr": "retail/orders/v2"},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Topic) {
				t.Helper()
				if beta.Spec.TopicString != "retail/orders/v2" {
					t.Fatalf("topicString = %q", beta.Spec.TopicString)
				}
			},
			assertBack: func(t *testing.T, orig, back *Topic) {
				t.Helper()
				if back.Spec.TopicString != "retail/orders/v2" {
					t.Fatalf("topicString = %q", back.Spec.TopicString)
				}
			},
		},
		{
			name: "folds publish subscribe and scope attributes",
			alpha: &Topic{
				ObjectMeta: testObjectMeta("t-access"),
				Spec: TopicSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					TopicName:     "RETAIL.ORDERS",
					Attributes: map[string]string{
						"pub":      "enabled",
						"sub":      "disabled",
						"pubscope": "ALL",
						"subscope": "QMGR",
						"defpsist": "yes",
						"descr":    "Retail topic",
					},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Topic) {
				t.Helper()
				if beta.Spec.Publish != messagingv1beta1.TopicAccessEnabled("enabled") {
					t.Fatalf("publish = %q", beta.Spec.Publish)
				}
				if beta.Spec.Subscribe != messagingv1beta1.TopicAccessEnabled("disabled") {
					t.Fatalf("subscribe = %q", beta.Spec.Subscribe)
				}
				if beta.Spec.PublishScope != "ALL" || beta.Spec.SubscribeScope != "QMGR" {
					t.Fatalf("scopes publish=%q subscribe=%q", beta.Spec.PublishScope, beta.Spec.SubscribeScope)
				}
				if beta.Spec.DefPersistence != messagingv1beta1.QueueDefaultPersistenceYes {
					t.Fatalf("defPersistence = %q", beta.Spec.DefPersistence)
				}
				if beta.Spec.Description != "Retail topic" {
					t.Fatalf("description = %q", beta.Spec.Description)
				}
				if len(beta.Spec.Attributes) != 0 {
					t.Fatalf("all promoted keys should be folded: %v", beta.Spec.Attributes)
				}
			},
			assertBack: func(t *testing.T, orig, back *Topic) {
				t.Helper()
				if back.Spec.PublishScope != "ALL" || back.Spec.Description != "Retail topic" {
					t.Fatalf("spec = %+v", back.Spec)
				}
			},
		},
		{
			name: "typed field wins over conflicting attribute",
			alpha: &Topic{
				ObjectMeta: testObjectMeta("t-prefer-typed"),
				Spec: TopicSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					TopicName:     "RETAIL.ORDERS",
					TopicString:   "typed/path",
					Attributes:    map[string]string{"topstr": "map/path"},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Topic) {
				t.Helper()
				if beta.Spec.TopicString != "typed/path" {
					t.Fatalf("topicString = %q", beta.Spec.TopicString)
				}
				if mapHasKey(beta.Spec.Attributes, "topstr") {
					t.Fatalf("conflicting attribute should be dropped")
				}
			},
			assertBack: func(t *testing.T, orig, back *Topic) {
				t.Helper()
				if back.Spec.TopicString != "typed/path" {
					t.Fatalf("topicString = %q", back.Spec.TopicString)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			beta, back := roundTripTopic(t, tt.alpha)
			if tt.assertBeta != nil {
				tt.assertBeta(t, beta)
			}
			if tt.assertBack != nil {
				tt.assertBack(t, tt.alpha, back)
			}
		})
	}
}

func TestTopicConvertFromHubRoundTrip(t *testing.T) {
	t.Parallel()

	beta := &messagingv1beta1.Topic{
		ObjectMeta: testObjectMeta("t-hub"),
		Spec: messagingv1beta1.TopicSpec{
			ConnectionRef:  messagingv1beta1.LocalObjectReference{Name: "qm1"},
			TopicName:      "RETAIL.HUB",
			TopicString:    "hub/topic",
			Publish:        messagingv1beta1.TopicAccessEnabledEnabled,
			DefPersistence: messagingv1beta1.QueueDefaultPersistenceNo,
			Attributes:     map[string]string{"custom": "y"},
		},
		Status: messagingv1beta1.TopicStatus{
			ObservedGeneration: 2,
			DesiredMQSC:        "DEFINE TOPIC(RETAIL.HUB)",
		},
	}

	_, back := roundTripBetaTopic(t, beta)
	if back.Spec.TopicString != "hub/topic" {
		t.Fatalf("topicString = %q", back.Spec.TopicString)
	}
	if back.Spec.Attributes["custom"] != "y" {
		t.Fatalf("attributes = %v", back.Spec.Attributes)
	}
	if back.Status.DesiredMQSC != "DEFINE TOPIC(RETAIL.HUB)" {
		t.Fatalf("desiredMQSC = %q", back.Status.DesiredMQSC)
	}
}
