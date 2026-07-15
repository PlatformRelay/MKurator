package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	"github.com/platformrelay/mkurator/internal/mqadmin"
	mqadmintest "github.com/platformrelay/mkurator/test/mocks/mqadmin"
)

var _ = Describe("ChannelAuthRuleReconciler", func() {
	const (
		ns          = "default"
		key         = "dev-app-addressmap"
		channelName = "DEV.APP.SVRCONN.0TLS"
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

	It("requeues when the connection is not Ready", func() {
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

		rule := sampleChannelAuthRule(ns, key, "qm1", channelName)
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())

		recorder := events.NewFakeRecorder(2)
		rec := &ChannelAuthRuleReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mqadmintest.NewMockFactory(GinkgoT()),
			Recorder:  recorder,
		}
		result, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(BeNumerically(">", 0))

		updated := &messagingv1alpha1.ChannelAuthRule{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(metav1.ConditionFalse))
		expectRecordedEvent(recorder, corev1.EventTypeNormal, messagingv1alpha1.ReasonProgressing)
	})

	It("applies CHLAUTH when the connection is Ready", func() {
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

		rule := sampleChannelAuthRule(ns, key, "qm1", channelName)
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())

		desired := mqadmin.ChannelAuthSpec{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			Address:     "*",
			UserSource:  "CHANNEL",
			CheckClient: "REQUIRED",
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannelAuth(mock.Anything, desired).Return(nil, mqadmin.ErrNotFound).Once()
		mockAdmin.EXPECT().SetChannelAuth(mock.Anything, desired).Return(nil).Once()

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().
			ForConnection(mock.Anything, mock.MatchedBy(func(c *messagingv1alpha1.QueueManagerConnection) bool {
				return c.Name == "qm1" && c.Namespace == ns
			})).
			Return(mockAdmin, nil)

		rec := &ChannelAuthRuleReconciler{
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

		updated := &messagingv1alpha1.ChannelAuthRule{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionStatus(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(metav1.ConditionTrue))
		Expect(updated.Status.MQObjectExists).NotTo(BeNil())
		Expect(*updated.Status.MQObjectExists).To(BeTrue())
		Expect(updated.Status.Message).To(Equal("ChannelAuthRule matches spec"))
		Expect(updated.Status.LastSyncTime).NotTo(BeNil())
	})

	It("applies BLOCKUSER CHLAUTH when the connection is Ready", func() {
		const blockKey = "dev-app-blockuser"

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

		rule := sampleBlockUserChannelAuthRule(ns, blockKey, "qm1", channelName)
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())

		desired := mqadmin.ChannelAuthSpec{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeBlockUser,
			UserList:    "nobody",
			Description: "Deny privileged user IDs",
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannelAuth(mock.Anything, desired).Return(nil, mqadmin.ErrNotFound).Once()
		mockAdmin.EXPECT().SetChannelAuth(mock.Anything, desired).Return(nil).Once()

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &ChannelAuthRuleReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: blockKey},
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: blockKey},
		})
		Expect(err).NotTo(HaveOccurred())

		updated := &messagingv1alpha1.ChannelAuthRule{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: blockKey}, updated)).To(Succeed())
		Expect(updated.Status.DesiredMQSC).To(ContainSubstring("TYPE(BLOCKUSER)"))
		Expect(updated.Status.DesiredMQSC).To(ContainSubstring("USERLIST('nobody')"))
	})

	It("skips SET when CHLAUTH already matches", func() {
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

		rule := sampleChannelAuthRule(ns, key, "qm1", channelName)
		rule.Finalizers = []string{messagingv1alpha1.ChannelAuthRuleFinalizer}
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())

		desired := mqadmin.ChannelAuthSpec{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			Address:     "*",
			UserSource:  "CHANNEL",
			CheckClient: "REQUIRED",
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannelAuth(mock.Anything, desired).Return(&mqadmin.ChannelAuthState{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			Address:     "*",
			UserSource:  "CHANNEL",
			CheckClient: "REQUIRED",
		}, nil).Once()

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &ChannelAuthRuleReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("emits a warning event when set channel auth fails terminally", func() {
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

		rule := sampleChannelAuthRule(ns, key, "qm1", channelName)
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())

		desired := mqadmin.ChannelAuthSpec{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			Address:     "*",
			UserSource:  "CHANNEL",
			CheckClient: "REQUIRED",
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannelAuth(mock.Anything, desired).Return(nil, mqadmin.ErrNotFound).Once()
		mockAdmin.EXPECT().SetChannelAuth(mock.Anything, desired).
			Return(&mqadmin.TerminalError{Reason: "MQSCError", Message: "set chlauth failed: AMQ8405E"})

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		recorder := events.NewFakeRecorder(2)
		rec := &ChannelAuthRuleReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
			Recorder:  recorder,
		}

		_, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		expectRecordedEvent(recorder, corev1.EventTypeWarning, "MQSCError")
	})

	It("requeues deletion when the connection is missing instead of failing terminally", func() {
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

		rule := sampleChannelAuthRule(ns, key, "qm1", channelName)
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, rule)).To(Succeed())
		controllerutil.AddFinalizer(rule, messagingv1alpha1.ChannelAuthRuleFinalizer)
		Expect(k8sClient.Update(ctx, rule)).To(Succeed())

		rec := &ChannelAuthRuleReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mqadmintest.NewMockFactory(GinkgoT()),
		}

		Expect(k8sClient.Delete(ctx, conn)).To(Succeed())
		Expect(k8sClient.Delete(ctx, rule)).To(Succeed())

		result, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(BeNumerically(">", 0))

		updated := &messagingv1alpha1.ChannelAuthRule{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(updated.DeletionTimestamp).NotTo(BeZero())
		Expect(conditionReason(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(messagingv1alpha1.ReasonProgressing))
	})

	It("reports drift without SET when observe-only is set", func() {
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

		rule := sampleChannelAuthRule(ns, key, "qm1", channelName)
		rule.Finalizers = []string{messagingv1alpha1.ChannelAuthRuleFinalizer}
		rule.Annotations = map[string]string{
			messagingv1alpha1.DriftPolicyAnnotation: messagingv1alpha1.DriftPolicyObserveOnly,
		}
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())

		desired := mqadmin.ChannelAuthSpec{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			Address:     "*",
			UserSource:  "CHANNEL",
			CheckClient: "REQUIRED",
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannelAuth(mock.Anything, desired).Return(&mqadmin.ChannelAuthState{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			Address:     "*",
			UserSource:  "CHANNEL",
			CheckClient: "ASQMGR",
		}, nil).Once()

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &ChannelAuthRuleReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		updated := &messagingv1alpha1.ChannelAuthRule{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionReason(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(messagingv1alpha1.ReasonDriftDetected))
	})

	It("reports not-found without SET when observe-only is set", func() {
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

		rule := sampleChannelAuthRule(ns, key, "qm1", channelName)
		rule.Finalizers = []string{messagingv1alpha1.ChannelAuthRuleFinalizer}
		rule.Annotations = map[string]string{
			messagingv1alpha1.DriftPolicyAnnotation: messagingv1alpha1.DriftPolicyObserveOnly,
		}
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())

		desired := mqadmin.ChannelAuthSpec{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			Address:     "*",
			UserSource:  "CHANNEL",
			CheckClient: "REQUIRED",
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannelAuth(mock.Anything, desired).Return(nil, mqadmin.ErrNotFound).Once()

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &ChannelAuthRuleReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		updated := &messagingv1alpha1.ChannelAuthRule{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: key}, updated)).To(Succeed())
		Expect(conditionReason(updated.Status.Conditions, messagingv1alpha1.ConditionSynced)).
			To(Equal(messagingv1alpha1.ReasonDriftDetected))
		Expect(updated.Status.MQObjectExists).NotTo(BeNil())
		Expect(*updated.Status.MQObjectExists).To(BeFalse())
	})

	It("removes CHLAUTH on delete", func() {
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

		rule := sampleChannelAuthRule(ns, key, "qm1", channelName)
		Expect(k8sClient.Create(ctx, rule)).To(Succeed())

		desired := mqadmin.ChannelAuthSpec{
			ChannelName: channelName,
			RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			Address:     "*",
			UserSource:  "CHANNEL",
			CheckClient: "REQUIRED",
		}

		mockAdmin := mqadmintest.NewMockAdmin(GinkgoT())
		mockAdmin.EXPECT().GetChannelAuth(mock.Anything, desired).Return(nil, mqadmin.ErrNotFound).Once()
		mockAdmin.EXPECT().SetChannelAuth(mock.Anything, desired).Return(nil).Once()
		mockAdmin.EXPECT().DeleteChannelAuth(mock.Anything, desired).Return(nil).Once()

		mockFactory := mqadmintest.NewMockFactory(GinkgoT())
		mockFactory.EXPECT().ForConnection(mock.Anything, mock.Anything).Return(mockAdmin, nil)

		rec := &ChannelAuthRuleReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			MQFactory: mockFactory,
		}

		_, err := rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		_, err = rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(k8sClient.Delete(ctx, rule)).To(Succeed())
		_, err = rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())
		_, err = rec.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: key},
		})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) {
			err := k8sClient.Get(
				ctx,
				types.NamespacedName{Namespace: ns, Name: key},
				&messagingv1alpha1.ChannelAuthRule{},
			)
			g.Expect(k8serrors.IsNotFound(err)).To(BeTrue())
		}).Should(Succeed())
	})
})

func sampleChannelAuthRule(ns, name, connName, channelName string) *messagingv1alpha1.ChannelAuthRule {
	return &messagingv1alpha1.ChannelAuthRule{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: connName},
			ChannelName:   channelName,
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
			UserSource:    "CHANNEL",
			CheckClient:   "REQUIRED",
		},
	}
}

func sampleBlockUserChannelAuthRule(ns, name, connName, channelName string) *messagingv1alpha1.ChannelAuthRule {
	return &messagingv1alpha1.ChannelAuthRule{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: connName},
			ChannelName:   channelName,
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeBlockUser,
			UserList:      "nobody",
			Description:   "Deny privileged user IDs",
		},
	}
}
