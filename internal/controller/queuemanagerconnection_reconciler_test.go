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
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konradheimel/kurator/api/v1alpha1"
	mqadmintest "github.com/konradheimel/kurator/test/mocks/mqadmin"
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
		cancel()
		_ = k8sClient.DeleteAllOf(ctx, &messagingv1alpha1.QueueManagerConnection{}, client.InNamespace(ns))
		_ = k8sClient.DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(ns))
	})

	It("sets Ready after a successful ping", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "mq-credentials", Namespace: ns},
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("secret"),
			},
		}
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		conn := &messagingv1alpha1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: key, Namespace: ns},
			Spec: messagingv1alpha1.QueueManagerConnectionSpec{
				QueueManager: "QM1",
				Endpoint:     "https://mq.example:9443",
				CredentialsSecretRef: messagingv1alpha1.SecretReference{
					Name: "mq-credentials",
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
})
