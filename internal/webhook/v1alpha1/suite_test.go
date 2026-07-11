package webhookv1alpha1

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
)

var (
	webhookTestEnv   *envtest.Environment
	webhookCfg       *rest.Config
	webhookK8sClient client.Client
	webhookCancel    context.CancelFunc
)

func TestWebhookAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook v1alpha1 envtest bootstrap")
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
	Expect(messagingv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
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
