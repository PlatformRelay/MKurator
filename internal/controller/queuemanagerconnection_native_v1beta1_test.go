package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
	mqadmintest "github.com/platformrelay/mkurator/test/mocks/mqadmin"
)

var _ = Describe("QueueManagerConnectionReconciler native v1beta1 (8e-1)", func() {
	const (
		ns  = "default"
		key = "qm-native"
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

	It("preserves spec.authentication across finalizer-add to Ready", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: ns},
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("secret"),
			},
		}
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		// Union-only v1beta1 spec: authentication.basic set, legacy credentialsSecretRef omitted.
		conn := &messagingv1beta1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: key, Namespace: ns},
			Spec: messagingv1beta1.QueueManagerConnectionSpec{
				QueueManager: testQueueManager,
				Endpoint:     testEndpoint,
				Authentication: &messagingv1beta1.MQWebAuthentication{
					Mode: messagingv1beta1.MQWebAuthenticationModeBasic,
					Basic: &messagingv1beta1.BasicAuth{
						SecretRef: messagingv1beta1.SecretReference{Name: testSecretName},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().Ping(mock.Anything).Return(nil).Maybe()

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil).Maybe()

		rec := &QueueManagerConnectionReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: key}}
		// Drive finalizer-add, then progressing + ping + ready. Each returns quickly.
		for i := 0; i < 3; i++ {
			_, err := rec.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
		}

		// AC: the stored hub still carries spec.authentication (no lossy round trip wiped it).
		stored := &messagingv1beta1.QueueManagerConnection{}
		Expect(k8sClient.Get(ctx, req.NamespacedName, stored)).To(Succeed())
		Expect(stored.Spec.Authentication).NotTo(BeNil(),
			"finalizer-add must not wipe spec.authentication (AUTH-14 data-loss root cause)")
		Expect(stored.Spec.Authentication.Mode).To(Equal(messagingv1beta1.MQWebAuthenticationModeBasic))
		Expect(stored.Spec.Authentication.Basic).NotTo(BeNil())
		Expect(stored.Spec.Authentication.Basic.SecretRef.Name).To(Equal(testSecretName))

		// The finalizer was added natively on the hub.
		Expect(stored.Finalizers).To(ContainElement(messagingv1beta1.QueueManagerConnectionFinalizer))
	})

	It("removes a leftover finalizer string on v1beta1 reconcile", func() {
		// A QMC persisted before the migration carries the finalizer constant. The native
		// v1beta1 reconcile matches and removes it on delete.
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: ns},
			Data:       map[string][]byte{"username": []byte("admin"), "password": []byte("secret")},
		}
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		conn := &messagingv1beta1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{
				Name:       key,
				Namespace:  ns,
				Finalizers: []string{messagingv1beta1.QueueManagerConnectionFinalizer},
			},
			Spec: messagingv1beta1.QueueManagerConnectionSpec{
				QueueManager: testQueueManager,
				Endpoint:     testEndpoint,
				Authentication: &messagingv1beta1.MQWebAuthentication{
					Mode: messagingv1beta1.MQWebAuthenticationModeBasic,
					Basic: &messagingv1beta1.BasicAuth{
						SecretRef: messagingv1beta1.SecretReference{Name: testSecretName},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())
		Expect(k8sClient.Delete(ctx, conn)).To(Succeed())

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ReleaseConnection(mock.Anything, mock.Anything).Return(nil).Once()

		rec := &QueueManagerConnectionReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}
		req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: key}}
		_, err := rec.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())

		got := &messagingv1beta1.QueueManagerConnection{}
		err = k8sClient.Get(ctx, req.NamespacedName, got)
		Expect(err).To(HaveOccurred(), "leftover finalizer removed → object deleted")
	})
})
