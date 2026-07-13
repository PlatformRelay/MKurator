package conversion

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
	conversionTestEnv *envtest.Environment
	conversionCfg     *rest.Config
)

func TestConversionSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Conversion webhook envtest bootstrap")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	conversionTestEnv = &envtest.Environment{
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
	conversionCfg, err = conversionTestEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(messagingv1beta1.AddToScheme(scheme.Scheme)).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(conversionTestEnv.Stop()).To(Succeed())
})

// newConversionManager builds a manager against the envtest config. The supplied
// scheme decides whether the v1beta1 hub kinds are registered, which lets tests
// drive SetupWithManager on both the happy path (types present) and the error
// path (types absent) without starting the manager.
func newConversionManager(sch *runtime.Scheme) (ctrl.Manager, error) {
	return ctrl.NewManager(conversionCfg, ctrl.Options{
		Scheme:  sch,
		Metrics: metricsserver.Options{BindAddress: "0"},
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    conversionTestEnv.WebhookInstallOptions.LocalServingHost,
			Port:    conversionTestEnv.WebhookInstallOptions.LocalServingPort,
			CertDir: conversionTestEnv.WebhookInstallOptions.LocalServingCertDir,
		}),
	})
}
