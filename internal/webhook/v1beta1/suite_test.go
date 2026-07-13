package webhookv1beta1

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

var (
	webhookTestEnv *envtest.Environment
	webhookCfg     *rest.Config
)

func TestWebhookAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook v1beta1 envtest bootstrap")
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
})

var _ = AfterSuite(func() {
	Expect(webhookTestEnv.Stop()).To(Succeed())
})

// newWebhookManager builds a manager against the envtest config. The supplied
// scheme decides whether the v1beta1 hub kinds are registered, which lets tests
// exercise both the happy path (types present) and the error path (types absent)
// without starting the manager or serving admission traffic.
func newWebhookManager(sch *runtime.Scheme) (ctrl.Manager, error) {
	return ctrl.NewManager(webhookCfg, ctrl.Options{
		Scheme:  sch,
		Metrics: metricsserver.Options{BindAddress: "0"},
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    webhookTestEnv.WebhookInstallOptions.LocalServingHost,
			Port:    webhookTestEnv.WebhookInstallOptions.LocalServingPort,
			CertDir: webhookTestEnv.WebhookInstallOptions.LocalServingCertDir,
		}),
	})
}
