package webhookv1alpha1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

func TestTopicWebhookValidateCreate(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := messagingv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
	v := &topicCustomValidator{Client: cl}

	topic := &messagingv1alpha1.Topic{
		ObjectMeta: metav1.ObjectMeta{Name: "t1", Namespace: "ns"},
		Spec: messagingv1alpha1.TopicSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			TopicName:     "RETAIL.ORDERS",
		},
	}
	if _, err := v.ValidateCreate(context.Background(), topic); err != nil {
		t.Fatalf("ValidateCreate: %v", err)
	}
}

func TestChannelWebhookValidateUpdate(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := messagingv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
	v := &channelCustomValidator{Client: cl}

	channel := &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: "c1", Namespace: "ns"},
		Spec: messagingv1alpha1.ChannelSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
		},
	}
	if _, err := v.ValidateUpdate(context.Background(), channel, channel); err != nil {
		t.Fatalf("ValidateUpdate: %v", err)
	}
}

func TestQueueManagerConnectionWebhookValidateDelete(t *testing.T) {
	v := &queueManagerConnectionCustomValidator{}
	warnings, err := v.ValidateDelete(context.Background(), &messagingv1alpha1.QueueManagerConnection{})
	if err != nil {
		t.Fatalf("ValidateDelete: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
}
