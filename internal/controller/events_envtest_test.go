package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
	"github.com/konih/kurator/internal/mqadmin"
	mqadmintest "github.com/konih/kurator/test/mocks/mqadmin"
)

var _ = Describe("events.k8s.io reconcile events", func() {
	const ns = "default"

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

	It("records Progressing on Queue when the connection is not Ready", func() {
		conn := &messagingv1alpha1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: ns},
			Spec: messagingv1alpha1.QueueManagerConnectionSpec{
				QueueManager: testQueueManager,
				Endpoint:     testEndpoint,
				CredentialsSecretRef: messagingv1alpha1.SecretReference{
					Name: testSecretName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())

		key := "orders-progressing"
		q := sampleQueue(ns, key, "qm1", testQueueName)
		Expect(k8sClient.Create(ctx, q)).To(Succeed())

		recorder := events.NewFakeRecorder(4)
		rec := &QueueReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mqadmintest.NewMockFactory(GinkgoT()),
			Recorder:  recorder,
		}
		_, err := rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		expectRecordedEvent(recorder, corev1.EventTypeNormal, messagingv1alpha1.ReasonProgressing)
	})

	It("records Available on Queue when reconcile succeeds", func() {
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

		key := "orders-available"
		q := sampleQueue(ns, key, "qm1", testQueueName)
		Expect(k8sClient.Create(ctx, q)).To(Succeed())

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().
			GetQueue(mock.Anything, mqadmin.QueueSpec{
				Name: testQueueName,
				Type: mqadmin.QueueTypeLocal,
				Attributes: map[string]string{
					testAttrMaxDepth: testMaxDepth,
				},
			}).
			Return(nil, &mqadmin.NotFoundError{Object: testQueueName})
		mockAdmin.EXPECT().
			DefineQueue(mock.Anything, mqadmin.QueueSpec{
				Name: testQueueName,
				Type: mqadmin.QueueTypeLocal,
				Attributes: map[string]string{
					testAttrMaxDepth: testMaxDepth,
				},
			}).
			Return(nil)

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		recorder := events.NewFakeRecorder(4)
		rec := &QueueReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  recorder,
		}

		_, err := rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		expectRecordedEvent(recorder, corev1.EventTypeNormal, messagingv1alpha1.ReasonAvailable)
	})

	It("records Available on Topic when reconcile succeeds", func() {
		const (
			key       = "retail-orders"
			topicName = "RETAIL.ORDERS"
		)

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

		topic := sampleTopic(ns, key, "qm1", topicName)
		Expect(k8sClient.Create(ctx, topic)).To(Succeed())

		desired := mqadmin.TopicSpec{
			Name: topicName,
			Attributes: map[string]string{
				"topstr": "retail/orders",
			},
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetTopic(mock.Anything, topicName).Return(nil, &mqadmin.NotFoundError{Object: topicName})
		mockAdmin.EXPECT().DefineTopic(mock.Anything, desired).Return(nil)

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		recorder := events.NewFakeRecorder(4)
		rec := &TopicReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  recorder,
		}

		_, err := rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		expectRecordedEvent(recorder, corev1.EventTypeNormal, messagingv1alpha1.ReasonAvailable)
	})
})
