package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
	"github.com/konih/mkurator/internal/mqadmin"
	mqadmintest "github.com/konih/mkurator/test/mocks/mqadmin"
)

var _ = Describe("ChannelReconciler", func() {
	const (
		ns          = "default"
		key         = "orders-app"
		channelName = "ORDERS.APP"
	)

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cleanupNamespace(context.Background(), ns)
		cancel()
	})

	It("defines the channel when the connection is Ready", func() {
		conn := readyConnection(ns, "qm1")
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())
		conn.Status = messagingv1alpha1.QueueManagerConnectionStatus{
			Conditions: []metav1.Condition{{
				Type:               messagingv1alpha1.ConditionReady,
				Status:             metav1.ConditionTrue,
				Reason:             messagingv1alpha1.ReasonAvailable,
				LastTransitionTime: metav1.Now(),
			}},
		}
		Expect(k8sClient.Status().Update(ctx, conn)).To(Succeed())

		channel := sampleChannel(ns, key, "qm1", channelName)
		Expect(k8sClient.Create(ctx, channel)).To(Succeed())

		desired := mqadmin.ChannelSpec{
			Name: channelName,
			Type: mqadmin.ChannelTypeSvrconn,
			Attributes: map[string]string{
				"trptype": "tcp",
			},
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannel(mock.Anything, desired).Return(nil, &mqadmin.NotFoundError{Object: channelName})
		mockAdmin.EXPECT().DefineChannel(mock.Anything, desired).Return(nil)

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &ChannelReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		expectDriftResyncRequeue(result)

		updated := &messagingv1alpha1.Channel{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(metav1.ConditionTrue))
	})
})

func sampleChannel(ns, name, connName, channelName string) *messagingv1alpha1.Channel {
	return &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: messagingv1alpha1.ChannelSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: connName},
			ChannelName:   channelName,
			Type:          messagingv1alpha1.ChannelTypeSvrconn,
			Attributes: map[string]string{
				"trptype": "tcp",
			},
		},
	}
}
