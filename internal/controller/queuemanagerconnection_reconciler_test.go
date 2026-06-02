package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
	"github.com/konih/kurator/internal/mqadmin"
	mqadmintest "github.com/konih/kurator/test/mocks/mqadmin"
)

var _ = Describe("QueueManagerConnectionReconciler", func() {
	const (
		ns  = "default"
		key = "qm1"
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

	It("sets Ready after a successful ping", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: ns},
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("secret"),
			},
		}
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		conn := &messagingv1alpha1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: key, Namespace: ns},
			Spec: messagingv1alpha1.QueueManagerConnectionSpec{
				QueueManager: testQueueManager,
				Endpoint:     testEndpoint,
				CredentialsSecretRef: messagingv1alpha1.SecretReference{
					Name: testSecretName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().Ping(mock.Anything).Return(nil)

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &QueueManagerConnectionReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		updated := &messagingv1alpha1.QueueManagerConnection{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionReady)).
			To(Equal(metav1.ConditionTrue))
		Expect(updated.Status.ObservedGeneration).To(Equal(updated.Generation))
	})

	It("sets Error when ping fails with a terminal error", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: ns},
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("secret"),
			},
		}
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		conn := &messagingv1alpha1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: key, Namespace: ns},
			Spec: messagingv1alpha1.QueueManagerConnectionSpec{
				QueueManager: testQueueManager,
				Endpoint:     testEndpoint,
				CredentialsSecretRef: messagingv1alpha1.SecretReference{
					Name: testSecretName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().Ping(mock.Anything).Return(&mqadmin.TerminalError{Message: "unauthorized"})

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &QueueManagerConnectionReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		updated := &messagingv1alpha1.QueueManagerConnection{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionReady)).
			To(Equal(metav1.ConditionFalse))
		Expect(conditionReason(updated.Status.Conditions, messagingv1alpha1.ConditionReady)).
			To(Equal(messagingv1alpha1.ReasonError))
	})

	It("removes the finalizer when the connection is deleted", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: ns},
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("secret"),
			},
		}
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		conn := &messagingv1alpha1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: key, Namespace: ns},
			Spec: messagingv1alpha1.QueueManagerConnectionSpec{
				QueueManager: testQueueManager,
				Endpoint:     testEndpoint,
				CredentialsSecretRef: messagingv1alpha1.SecretReference{
					Name: testSecretName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, conn)).To(Succeed())

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		rec := &QueueManagerConnectionReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(k8sClient.Delete(ctx, conn)).To(Succeed())

		mockFactory.EXPECT().ReleaseConnection(mock.Anything, mock.Anything).Return(nil).Once()

		_, err = rec.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		updated := &messagingv1alpha1.QueueManagerConnection{}
		err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)
		Expect(apierrors.IsNotFound(err)).To(BeTrue(), "finalizer removed and object deleted from API")
	})
})
