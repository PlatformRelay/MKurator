package webhookv1alpha1

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

func TestTopicWebhookValidateCreate(t *testing.T) {
	RegisterTestingT(t)
	scheme := runtime.NewScheme()
	Expect(messagingv1alpha1.AddToScheme(scheme)).To(Succeed())

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
	_, err := v.ValidateCreate(context.Background(), topic)
	Expect(err).NotTo(HaveOccurred())
}

func TestChannelWebhookValidateUpdate(t *testing.T) {
	RegisterTestingT(t)
	scheme := runtime.NewScheme()
	Expect(messagingv1alpha1.AddToScheme(scheme)).To(Succeed())

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
	_, err := v.ValidateUpdate(context.Background(), channel, channel)
	Expect(err).NotTo(HaveOccurred())
}

func TestQueueManagerConnectionWebhookValidateDelete(t *testing.T) {
	RegisterTestingT(t)
	v := &queueManagerConnectionCustomValidator{}
	warnings, err := v.ValidateDelete(context.Background(), &messagingv1alpha1.QueueManagerConnection{})
	Expect(err).NotTo(HaveOccurred())
	Expect(warnings).To(BeEmpty())
}

func TestSetupWithManagerRegistersWebhooks(t *testing.T) {
	RegisterTestingT(t)
	Expect(admissionResult).NotTo(BeNil())
}
