package webhookv1beta1

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

var (
	webhookTestEnv   *envtest.Environment
	webhookCfg       *rest.Config
	webhookK8sClient client.Client
	webhookCancel    context.CancelFunc
)

func TestWebhookAdmissionV1Beta1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Admission v1beta1 Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	webhookTestEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crd", "bases"),
		},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{
				filepath.Join("..", "..", "..", "config", "webhook", "manifests.yaml"),
			},
		},
	}

	var err error
	webhookCfg, err = webhookTestEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(messagingv1beta1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(corev1.AddToScheme(scheme.Scheme)).To(Succeed())

	webhookK8sClient, err = client.New(webhookCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	mgr, err := ctrl.NewManager(webhookCfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    webhookTestEnv.WebhookInstallOptions.LocalServingHost,
			Port:    webhookTestEnv.WebhookInstallOptions.LocalServingPort,
			CertDir: webhookTestEnv.WebhookInstallOptions.LocalServingCertDir,
		}),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(SetupWithManager(mgr)).To(Succeed())

	ctx, cancel := context.WithCancel(context.Background())
	webhookCancel = cancel
	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()
	time.Sleep(2 * time.Second)
})

var _ = AfterSuite(func() {
	if webhookCancel != nil {
		webhookCancel()
	}
	Expect(webhookTestEnv.Stop()).To(Succeed())
})

var _ = Describe("v1beta1 validating admission", func() {
	const ns = "webhook-v1beta1-test"

	BeforeEach(func() {
		ctx := context.Background()
		nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
		_ = webhookK8sClient.Create(ctx, nsObj)
		cleanupWebhookNamespaceV1Beta1(ctx, ns)
	})

	AfterEach(func() {
		cleanupWebhookNamespaceV1Beta1(context.Background(), ns)
	})

	It("denies Queue when connectionRef target is missing", func() {
		ctx := context.Background()
		q := &messagingv1beta1.Queue{
			ObjectMeta: metav1.ObjectMeta{Name: "bad-queue-v1beta1", Namespace: ns},
			Spec: messagingv1beta1.QueueSpec{
				ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "missing-qmc"},
				QueueName:     "APP.ORDERS",
			},
		}
		err := webhookK8sClient.Create(ctx, q)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsInvalid(err)).To(BeTrue())
	})

	It("allows valid Topic and Channel when connection exists", func() {
		ctx := context.Background()
		Expect(webhookK8sClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: ns},
		})).To(Succeed())
		Expect(webhookK8sClient.Create(ctx, sampleWebhookConnectionV1Beta1(ns, "qm-beta"))).To(Succeed())

		topic := &messagingv1beta1.Topic{
			ObjectMeta: metav1.ObjectMeta{Name: "good-topic-v1beta1", Namespace: ns},
			Spec: messagingv1beta1.TopicSpec{
				ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm-beta"},
				TopicName:     "RETAIL.ORDERS",
			},
		}
		Expect(webhookK8sClient.Create(ctx, topic)).To(Succeed())

		channel := &messagingv1beta1.Channel{
			ObjectMeta: metav1.ObjectMeta{Name: "good-channel-v1beta1", Namespace: ns},
			Spec: messagingv1beta1.ChannelSpec{
				ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm-beta"},
				ChannelName:   "ORDERS.APP",
			},
		}
		Expect(webhookK8sClient.Create(ctx, channel)).To(Succeed())
	})

	It("allows valid ChannelAuthRule and AuthorityRecord", func() {
		ctx := context.Background()
		Expect(webhookK8sClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: ns},
		})).To(Succeed())
		Expect(webhookK8sClient.Create(ctx, sampleWebhookConnectionV1Beta1(ns, "qm-beta-auth"))).To(Succeed())
		Expect(webhookK8sClient.Create(ctx, &messagingv1beta1.Channel{
			ObjectMeta: metav1.ObjectMeta{Name: "orders-app-v1beta1", Namespace: ns},
			Spec: messagingv1beta1.ChannelSpec{
				ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm-beta-auth"},
				ChannelName:   "ORDERS.APP",
			},
		})).To(Succeed())

		rule := &messagingv1beta1.ChannelAuthRule{
			ObjectMeta: metav1.ObjectMeta{Name: "good-car-v1beta1", Namespace: ns},
			Spec: messagingv1beta1.ChannelAuthRuleSpec{
				ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm-beta-auth"},
				ChannelName:   "ORDERS.APP",
				RuleType:      messagingv1beta1.ChannelAuthRuleTypeAddressMap,
				Address:       "*",
			},
		}
		Expect(webhookK8sClient.Create(ctx, rule)).To(Succeed())

		auth := &messagingv1beta1.AuthorityRecord{
			ObjectMeta: metav1.ObjectMeta{Name: "good-auth-v1beta1", Namespace: ns},
			Spec: messagingv1beta1.AuthorityRecordSpec{
				ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm-beta-auth"},
				Profile:       "APP.ORDERS",
				ObjectType:    messagingv1beta1.AuthorityObjectTypeQueue,
				Principal:     "app",
				Authorities:   []string{"GET"},
			},
		}
		Expect(webhookK8sClient.Create(ctx, auth)).To(Succeed())
	})

	It("denies QueueManagerConnection delete when dependents exist", func() {
		ctx := context.Background()
		Expect(webhookK8sClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: ns},
		})).To(Succeed())
		conn := sampleWebhookConnectionV1Beta1(ns, "qm-beta-delete")
		Expect(webhookK8sClient.Create(ctx, conn)).To(Succeed())
		Expect(webhookK8sClient.Create(ctx, &messagingv1beta1.Queue{
			ObjectMeta: metav1.ObjectMeta{Name: "dep-queue-v1beta1", Namespace: ns},
			Spec: messagingv1beta1.QueueSpec{
				ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm-beta-delete"},
				QueueName:     "APP.ORDERS",
			},
		})).To(Succeed())

		err := webhookK8sClient.Delete(ctx, conn)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsInvalid(err)).To(BeTrue())
	})

	Describe("deprecated spec.attributes warnings", func() {
		BeforeEach(func() {
			ctx := context.Background()
			Expect(webhookK8sClient.Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: ns},
			})).To(Succeed())
			Expect(webhookK8sClient.Create(ctx, sampleWebhookConnectionV1Beta1(ns, "qm-v1beta1"))).To(Succeed())
		})

		It("allows Queue create and warns on deprecated attribute keys", func() {
			ctx := context.Background()
			warningClient, capture := newWarningCapturingClientV1Beta1()
			q := &messagingv1beta1.Queue{
				ObjectMeta: metav1.ObjectMeta{Name: "warn-queue-v1beta1", Namespace: ns},
				Spec: messagingv1beta1.QueueSpec{
					ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm-v1beta1"},
					QueueName:     "APP.WARN.BETA",
					Attributes:    map[string]string{"maxdepth": "1000"},
				},
			}
			Expect(warningClient.Create(ctx, q)).To(Succeed())
			expectDeprecatedAttributeWarningV1Beta1(capture, "maxdepth", "spec.maxDepth")
		})
	})
})

func sampleWebhookConnectionV1Beta1(ns, name string) *messagingv1beta1.QueueManagerConnection {
	return &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{
				Name: "creds",
			},
		},
	}
}

func cleanupWebhookNamespaceV1Beta1(ctx context.Context, ns string) {
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1beta1.Queue{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1beta1.Topic{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1beta1.ChannelAuthRule{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1beta1.AuthorityRecord{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1beta1.Channel{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1beta1.QueueManagerConnection{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(ns))
}

type warningCaptureV1Beta1 struct {
	store *[]string
	mu    *sync.Mutex
}

func (w warningCaptureV1Beta1) HandleWarningHeader(_ int, _ string, text string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	*w.store = append(*w.store, text)
}

func newWarningCapturingClientV1Beta1() (client.Client, *warningCaptureV1Beta1) {
	var (
		mu       sync.Mutex
		warnings []string
	)
	capture := &warningCaptureV1Beta1{store: &warnings, mu: &mu}
	warningCfg := rest.CopyConfig(webhookCfg)
	warningCfg.WarningHandler = capture
	warningClient, err := client.New(warningCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	return warningClient, capture
}

func expectDeprecatedAttributeWarningV1Beta1(capture *warningCaptureV1Beta1, attrKey, replacement string) {
	capture.mu.Lock()
	defer capture.mu.Unlock()
	combined := strings.Join(*capture.store, " ")
	Expect(*capture.store).NotTo(BeEmpty(), "expected admission warnings, got none")
	Expect(combined).To(ContainSubstring(attrKey))
	Expect(combined).To(ContainSubstring(replacement))
	Expect(combined).To(ContainSubstring("is deprecated"))
}
