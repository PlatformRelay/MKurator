package webhookv1alpha1

import (
	"context"
	"path/filepath"
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
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

var (
	webhookTestEnv   *envtest.Environment
	webhookCfg       *rest.Config
	webhookK8sClient client.Client
	webhookCancel    context.CancelFunc
)

func TestWebhookAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Admission Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping webhook test environment")
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
	Expect(messagingv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(corev1.AddToScheme(scheme.Scheme)).To(Succeed())

	webhookK8sClient, err = client.New(webhookCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	mgr, err := ctrl.NewManager(webhookCfg, ctrl.Options{
		Scheme: scheme.Scheme,
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

var _ = Describe("Validating admission webhooks", func() {
	const ns = "webhook-admission"

	BeforeEach(func() {
		ctx := context.Background()
		nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
		_ = webhookK8sClient.Create(ctx, nsObj)
		cleanupWebhookNamespace(ctx, ns)
	})

	AfterEach(func() {
		cleanupWebhookNamespace(context.Background(), ns)
	})

	It("denies Queue when connectionRef target is missing", func() {
		ctx := context.Background()
		q := &messagingv1alpha1.Queue{
			ObjectMeta: metav1.ObjectMeta{Name: "bad-queue", Namespace: ns},
			Spec: messagingv1alpha1.QueueSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "missing-qmc"},
				QueueName:     "APP.ORDERS",
			},
		}
		err := webhookK8sClient.Create(ctx, q)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsInvalid(err)).To(BeTrue())
	})

	It("denies alias Queue without targq", func() {
		ctx := context.Background()
		Expect(webhookK8sClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: ns},
		})).To(Succeed())
		conn := sampleWebhookConnection(ns, "qm1")
		Expect(webhookK8sClient.Create(ctx, conn)).To(Succeed())

		q := &messagingv1alpha1.Queue{
			ObjectMeta: metav1.ObjectMeta{Name: "alias-queue", Namespace: ns},
			Spec: messagingv1alpha1.QueueSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				QueueName:     "ALIAS.Q",
				Type:          messagingv1alpha1.QueueTypeAlias,
			},
		}
		err := webhookK8sClient.Create(ctx, q)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsInvalid(err)).To(BeTrue())
	})

	It("allows valid Queue when connection and spec are valid", func() {
		ctx := context.Background()
		Expect(webhookK8sClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: ns},
		})).To(Succeed())
		conn := sampleWebhookConnection(ns, "qm1")
		Expect(webhookK8sClient.Create(ctx, conn)).To(Succeed())

		q := &messagingv1alpha1.Queue{
			ObjectMeta: metav1.ObjectMeta{Name: "good-queue", Namespace: ns},
			Spec: messagingv1alpha1.QueueSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				QueueName:     "APP.ORDERS",
				Attributes:    map[string]string{"maxdepth": "1000"},
			},
		}
		Expect(webhookK8sClient.Create(ctx, q)).To(Succeed())
	})
})

func sampleWebhookConnection(ns, name string) *messagingv1alpha1.QueueManagerConnection {
	return &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{
				Name: "creds",
			},
		},
	}
}

func cleanupWebhookNamespace(ctx context.Context, ns string) {
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.Queue{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.Topic{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.Channel{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.QueueManagerConnection{}, client.InNamespace(ns))
	_ = webhookK8sClient.DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(ns))
}
