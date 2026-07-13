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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	"github.com/platformrelay/mkurator/internal/mqadmin"
	mqadmintest "github.com/platformrelay/mkurator/test/mocks/mqadmin"
)

// These suites pin the error/requeue branches of the MQ-object reconcilers that
// touch IBM MQ: a missing connection, an mqweb factory error, a transient GetX
// failure (requeue), and an adoption-policy block. They complement the happy-path
// and terminal-define suites already present in the *_reconciler_test.go files.

var _ = Describe("TopicReconciler error branches", func() {
	const (
		ns        = "default"
		key       = "retail-orders"
		topicName = "RETAIL.ORDERS"
	)

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() { ctx, cancel = context.WithCancel(context.Background()) })
	AfterEach(func() {
		cleanupNamespace(context.Background(), ns)
		cancel()
	})

	It("sets Synced=False when the referenced connection does not exist", func() {
		topic := sampleTopic(ns, key, "missing-qm", topicName)
		Expect(k8sClient.Create(ctx, topic)).To(Succeed())

		recorder := events.NewFakeRecorder(2)
		rec := &TopicReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mqadmintest.NewMockFactory(GinkgoT()),
			Recorder:  recorder,
		}

		result, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))

		updated := &messagingv1alpha1.Topic{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(metav1.ConditionFalse))
	})

	It("requeues when the mqweb factory fails transiently", func() {
		readyConnectionStatus(ctx, ns, "qm1")

		topic := sampleTopic(ns, key, "qm1", topicName)
		Expect(k8sClient.Create(ctx, topic)).To(Succeed())

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).
			Return(nil, &mqadmin.TransientError{Message: "mqweb unreachable"})

		recorder := events.NewFakeRecorder(2)
		rec := &TopicReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  recorder,
		}

		result, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(Equal(TransientRequeueInterval()))

		updated := &messagingv1alpha1.Topic{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(metav1.ConditionFalse))
	})

	It("blocks adoption when FailIfExists and the topic already exists", func() {
		readyConnectionStatus(ctx, ns, "qm1")

		topic := sampleTopic(ns, key, "qm1", topicName)
		topic.Spec.AdoptionPolicy = messagingv1alpha1.AdoptionPolicyFailIfExists
		Expect(k8sClient.Create(ctx, topic)).To(Succeed())

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetTopic(mock.Anything, topicName).
			Return(&mqadmin.TopicState{Name: topicName, Attributes: map[string]string{"topstr": "retail/orders"}}, nil)

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		recorder := events.NewFakeRecorder(2)
		rec := &TopicReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  recorder,
		}

		// First reconcile adds the finalizer.
		_, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		expectDriftResyncRequeue(result)

		updated := &messagingv1alpha1.Topic{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(metav1.ConditionFalse))
		Expect(conditionReason(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(messagingv1alpha1.ReasonAlreadyExists))
	})
})

var _ = Describe("ChannelReconciler error branches", func() {
	const (
		ns          = "default"
		key         = "orders-app"
		channelName = "ORDERS.APP"
	)

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() { ctx, cancel = context.WithCancel(context.Background()) })
	AfterEach(func() {
		cleanupNamespace(context.Background(), ns)
		cancel()
	})

	It("sets Synced=False when the referenced connection does not exist", func() {
		channel := sampleChannel(ns, key, "missing-qm", channelName)
		Expect(k8sClient.Create(ctx, channel)).To(Succeed())

		recorder := events.NewFakeRecorder(2)
		rec := &ChannelReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mqadmintest.NewMockFactory(GinkgoT()),
			Recorder:  recorder,
		}

		result, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))

		updated := &messagingv1alpha1.Channel{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(metav1.ConditionFalse))
	})

	It("requeues when the mqweb factory fails transiently", func() {
		readyConnectionStatus(ctx, ns, "qm1")

		channel := sampleChannel(ns, key, "qm1", channelName)
		Expect(k8sClient.Create(ctx, channel)).To(Succeed())

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).
			Return(nil, &mqadmin.TransientError{Message: "mqweb unreachable"})

		rec := &ChannelReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  events.NewFakeRecorder(2),
		}

		result, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(Equal(TransientRequeueInterval()))
	})

	It("blocks adoption when FailIfExists and the channel already exists", func() {
		readyConnectionStatus(ctx, ns, "qm1")

		channel := sampleChannel(ns, key, "qm1", channelName)
		channel.Spec.AdoptionPolicy = messagingv1alpha1.AdoptionPolicyFailIfExists
		Expect(k8sClient.Create(ctx, channel)).To(Succeed())

		desired := mqadmin.ChannelSpec{
			Name:       channelName,
			Type:       mqadmin.ChannelTypeSvrconn,
			Attributes: map[string]string{"trptype": "tcp"},
		}
		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannel(mock.Anything, desired).
			Return(&mqadmin.ChannelState{Name: channelName, Attributes: map[string]string{"trptype": "tcp"}}, nil)

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		recorder := events.NewFakeRecorder(2)
		rec := &ChannelReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  recorder,
		}

		// First reconcile adds the finalizer.
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
			To(Equal(metav1.ConditionFalse))
		Expect(conditionReason(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(messagingv1alpha1.ReasonAlreadyExists))
	})
})

var _ = Describe("QueueManagerConnectionReconciler error branches", func() {
	const (
		ns  = "default"
		key = "qm1"
	)

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() { ctx, cancel = context.WithCancel(context.Background()) })
	AfterEach(func() {
		cleanupNamespace(context.Background(), ns)
		cancel()
	})

	It("requeues when building the mqweb client fails transiently", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: ns},
			Data:       map[string][]byte{"username": []byte("admin"), "password": []byte("secret")},
		}
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		conn := &messagingv1alpha1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: key, Namespace: ns},
			Spec: messagingv1alpha1.QueueManagerConnectionSpec{
				QueueManager:         testQueueManager,
				Endpoint:             testEndpoint,
				CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: testSecretName},
			},
		}
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).
			Return(nil, &mqadmin.TransientError{Message: "dial mqweb"})

		recorder := events.NewFakeRecorder(2)
		rec := &QueueManagerConnectionReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  recorder,
		}
		req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: key}}

		// First reconcile adds the finalizer.
		_, err := rec.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())

		result, err := rec.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(Equal(TransientRequeueInterval()))

		updated := &messagingv1alpha1.QueueManagerConnection{}
		Expect(k8sClient.Get(ctx, req.NamespacedName, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionReady)).
			To(Equal(metav1.ConditionFalse))
		Expect(conditionReason(updated.Status.Conditions, messagingv1alpha1.ConditionReady)).
			To(Equal(messagingv1alpha1.ReasonError))
	})

	It("requeues on a transient ping failure when not yet Ready", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: ns},
			Data:       map[string][]byte{"username": []byte("admin"), "password": []byte("secret")},
		}
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		conn := &messagingv1alpha1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: key, Namespace: ns},
			Spec: messagingv1alpha1.QueueManagerConnectionSpec{
				QueueManager:         testQueueManager,
				Endpoint:             testEndpoint,
				CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: testSecretName},
			},
		}
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().Ping(mock.Anything).Return(&mqadmin.TransientError{Message: "timeout"})
		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &QueueManagerConnectionReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  events.NewFakeRecorder(2),
		}
		req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: key}}

		// First reconcile adds the finalizer.
		_, err := rec.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())

		result, err := rec.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(Equal(TransientRequeueInterval()))

		updated := &messagingv1alpha1.QueueManagerConnection{}
		Expect(k8sClient.Get(ctx, req.NamespacedName, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionReady)).
			To(Equal(metav1.ConditionFalse))
	})
})

// readyConnectionStatus creates a QueueManagerConnection and marks it Ready so
// MQ-object reconciles proceed past the wait-for-connection gate.
func readyConnectionStatus(
	ctx context.Context,
	ns, name string,
) *messagingv1alpha1.QueueManagerConnection {
	GinkgoHelper()
	conn := readyConnection(ns, name)
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
	return conn
}
