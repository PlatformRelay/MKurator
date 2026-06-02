package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/konradheimel/kurator/api/v1alpha1"
)

func TestRequestsForConnection_EnqueuesDependents(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "kurator-system"
	s := runtime.NewScheme()
	if err := messagingv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: ns},
	}
	queue := &messagingv1alpha1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: ns},
		Spec: messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.ORDERS",
		},
	}
	topic := &messagingv1alpha1.Topic{
		ObjectMeta: metav1.ObjectMeta{Name: "retail", Namespace: ns},
		Spec: messagingv1alpha1.TopicSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			TopicName:     "RETAIL.ORDERS",
		},
	}
	channel := &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: ns},
		Spec: messagingv1alpha1.ChannelSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(conn, queue, topic, channel).Build()
	reqs := requestsForConnection(ctx, cl, conn)
	if len(reqs) != 3 {
		t.Fatalf("requests = %d, want 3", len(reqs))
	}
}

func TestConnectionRefName(t *testing.T) {
	t.Parallel()
	q := &messagingv1alpha1.Queue{
		Spec: messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
		},
	}
	name, err := connectionRefName(q)
	if err != nil || name != "qm1" {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestConnectionReadyChanged(t *testing.T) {
	t.Parallel()
	old := &messagingv1alpha1.QueueManagerConnection{}
	newReady := &messagingv1alpha1.QueueManagerConnection{
		Status: messagingv1alpha1.QueueManagerConnectionStatus{
			Conditions: []metav1.Condition{{
				Type:   messagingv1alpha1.ConditionReady,
				Status: metav1.ConditionTrue,
			}},
		},
	}
	if !connectionReadyChanged(old, newReady) {
		t.Fatal("expected ready transition")
	}
}
