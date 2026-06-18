package conversionwebhook_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	messagingv1alpha1 "github.com/conduit-ops/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
	webhookconversion "github.com/conduit-ops/mkurator/internal/webhook/conversion"
	webhookv1alpha1 "github.com/conduit-ops/mkurator/internal/webhook/v1alpha1"
)

var (
	conversionTestEnv   *envtest.Environment
	conversionCfg       *rest.Config
	conversionK8sClient client.Client
	conversionScheme    *runtime.Scheme
	conversionCancel    context.CancelFunc
)

func TestConversionWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Conversion Webhook Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	root, err := filepath.Abs(filepath.Join("..", ".."))
	Expect(err).NotTo(HaveOccurred())
	crdDir, err := buildCRDInstallDir(root)
	Expect(err).NotTo(HaveOccurred())

	conversionScheme = runtime.NewScheme()
	Expect(messagingv1alpha1.AddToScheme(conversionScheme)).To(Succeed())
	Expect(messagingv1beta1.AddToScheme(conversionScheme)).To(Succeed())
	Expect(corev1.AddToScheme(conversionScheme)).To(Succeed())

	By("bootstrapping conversion webhook test environment")
	conversionTestEnv = &envtest.Environment{
		Scheme:            conversionScheme,
		CRDDirectoryPaths: []string{crdDir},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{
				filepath.Join(root, "config", "webhook", "manifests.yaml"),
			},
		},
	}

	conversionCfg, err = conversionTestEnv.Start()
	Expect(err).NotTo(HaveOccurred())

	conversionK8sClient, err = client.New(conversionCfg, client.Options{Scheme: conversionScheme})
	Expect(err).NotTo(HaveOccurred())

	mgr, err := ctrl.NewManager(conversionCfg, ctrl.Options{
		Scheme:  conversionScheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    conversionTestEnv.WebhookInstallOptions.LocalServingHost,
			Port:    conversionTestEnv.WebhookInstallOptions.LocalServingPort,
			CertDir: conversionTestEnv.WebhookInstallOptions.LocalServingCertDir,
		}),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(webhookconversion.SetupWithManager(mgr)).To(Succeed())
	Expect(webhookv1alpha1.SetupWithManager(mgr)).To(Succeed())

	ctx, cancel := context.WithCancel(context.Background())
	conversionCancel = cancel
	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()
	time.Sleep(2 * time.Second)
})

var _ = AfterSuite(func() {
	if conversionCancel != nil {
		conversionCancel()
	}
	if conversionTestEnv != nil {
		Expect(conversionTestEnv.Stop()).To(Succeed())
	}
})

var _ = Describe("CRD conversion round-trip", func() {
	const ns = "conversion-webhook"

	BeforeEach(func() {
		ctx := context.Background()
		err := conversionK8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
		Expect(client.IgnoreAlreadyExists(err)).To(Succeed())
		Expect(conversionFixture(ctx, ns)).To(Succeed())
	})

	AfterEach(func() {
		ctx := context.Background()
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.Queue{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1beta1.Queue{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.Topic{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1beta1.Topic{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.Channel{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1beta1.Channel{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.ChannelAuthRule{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1beta1.ChannelAuthRule{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1alpha1.AuthorityRecord{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(ctx, &messagingv1beta1.AuthorityRecord{}, client.InNamespace(ns))
		_ = conversionK8sClient.DeleteAllOf(
			ctx,
			&messagingv1alpha1.QueueManagerConnection{},
			client.InNamespace(ns),
		)
		_ = conversionK8sClient.DeleteAllOf(
			ctx,
			&messagingv1beta1.QueueManagerConnection{},
			client.InNamespace(ns),
		)
	})

	It("round-trips Queue v1alpha1 through v1beta1 and back", func() {
		ctx := context.Background()
		alpha := &messagingv1alpha1.Queue{
			ObjectMeta: metav1.ObjectMeta{Name: "rt-queue", Namespace: ns},
			Spec: messagingv1alpha1.QueueSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				QueueName:     "APP.ORDERS",
				Attributes: map[string]string{
					"maxdepth": "5000",
					"custom":   "keep-me",
				},
			},
			Status: messagingv1alpha1.QueueStatus{},
		}
		Expect(conversionK8sClient.Create(ctx, alpha)).To(Succeed())
		alpha.Status.ObservedGeneration = 2
		alpha.Status.DesiredMQSC = "DEFINE QLOCAL(APP.ORDERS) REPLACE MAXDEPTH(5000)"
		Expect(conversionK8sClient.Status().Update(ctx, alpha)).To(Succeed())

		beta := &messagingv1beta1.Queue{}
		Expect(conversionK8sClient.Get(ctx, client.ObjectKeyFromObject(alpha), beta)).To(Succeed())
		Expect(beta.Spec.MaxDepth).NotTo(BeNil())
		Expect(*beta.Spec.MaxDepth).To(Equal(int32(5000)))
		Expect(beta.Spec.Attributes).To(HaveKeyWithValue("custom", "keep-me"))
		Expect(beta.Spec.Attributes).NotTo(HaveKey("maxdepth"))
		Expect(beta.Status.ObservedGeneration).To(Equal(int64(2)))
		Expect(beta.Status.DesiredMQSC).To(Equal("DEFINE QLOCAL(APP.ORDERS) REPLACE MAXDEPTH(5000)"))

		roundTrip := &messagingv1alpha1.Queue{}
		Expect(conversionK8sClient.Get(ctx, client.ObjectKeyFromObject(alpha), roundTrip)).To(Succeed())
		Expect(roundTrip.Spec.Attributes).To(HaveKeyWithValue("custom", "keep-me"))
		Expect(roundTrip.Spec.Attributes).To(HaveKeyWithValue("maxdepth", "5000"))
	})

	It("round-trips Topic v1alpha1 attribute folding through v1beta1", func() {
		ctx := context.Background()
		alpha := &messagingv1alpha1.Topic{
			ObjectMeta: metav1.ObjectMeta{Name: "rt-topic", Namespace: ns},
			Spec: messagingv1alpha1.TopicSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				TopicName:     "RETAIL.ORDERS",
				Attributes:    map[string]string{"topstr": "retail/orders", "extra": "x"},
			},
		}
		Expect(conversionK8sClient.Create(ctx, alpha)).To(Succeed())

		beta := &messagingv1beta1.Topic{}
		Expect(conversionK8sClient.Get(ctx, client.ObjectKeyFromObject(alpha), beta)).To(Succeed())
		Expect(beta.Spec.TopicString).To(Equal("retail/orders"))
		Expect(beta.Spec.Attributes).To(HaveKeyWithValue("extra", "x"))
		Expect(beta.Spec.Attributes).NotTo(HaveKey("topstr"))
	})

	It("round-trips Channel v1alpha1 through v1beta1", func() {
		ctx := context.Background()
		alpha := &messagingv1alpha1.Channel{
			ObjectMeta: metav1.ObjectMeta{Name: "rt-channel", Namespace: ns},
			Spec: messagingv1alpha1.ChannelSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				ChannelName:   "ORDERS.APP",
				Attributes:    map[string]string{"trptype": "tcp", "note": "y"},
			},
		}
		Expect(conversionK8sClient.Create(ctx, alpha)).To(Succeed())

		beta := &messagingv1beta1.Channel{}
		Expect(conversionK8sClient.Get(ctx, client.ObjectKeyFromObject(alpha), beta)).To(Succeed())
		Expect(beta.Spec.TransportType).To(Equal(messagingv1beta1.ChannelTransportType("tcp")))
		Expect(beta.Spec.Attributes).To(HaveKeyWithValue("note", "y"))
	})

	It("round-trips ChannelAuthRule v1alpha1 through v1beta1", func() {
		ctx := context.Background()
		alpha := &messagingv1alpha1.ChannelAuthRule{
			ObjectMeta: metav1.ObjectMeta{Name: "rt-car", Namespace: ns},
			Spec: messagingv1alpha1.ChannelAuthRuleSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				ChannelName:   "ORDERS.APP",
				RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
				Address:       "*",
				UserSource:    messagingv1alpha1.ChannelAuthUserSourceMap,
				McaUser:       "app",
			},
			Status: messagingv1alpha1.ChannelAuthRuleStatus{DesiredMQSC: "SET CHLAUTH ..."},
		}
		Expect(conversionK8sClient.Create(ctx, alpha)).To(Succeed())

		beta := &messagingv1beta1.ChannelAuthRule{}
		Expect(conversionK8sClient.Get(ctx, client.ObjectKeyFromObject(alpha), beta)).To(Succeed())
		Expect(beta.Spec.McaUser).To(Equal("app"))
		Expect(beta.Status.DesiredMQSC).To(Equal(alpha.Status.DesiredMQSC))
	})

	It("round-trips AuthorityRecord v1alpha1 through v1beta1", func() {
		ctx := context.Background()
		alpha := &messagingv1alpha1.AuthorityRecord{
			ObjectMeta: metav1.ObjectMeta{Name: "rt-auth", Namespace: ns},
			Spec: messagingv1alpha1.AuthorityRecordSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				Profile:       "APP.ORDERS",
				ObjectType:    messagingv1alpha1.AuthorityObjectTypeQueue,
				Principal:     "app",
				Authorities:   []string{"GET", "PUT"},
			},
		}
		Expect(conversionK8sClient.Create(ctx, alpha)).To(Succeed())

		beta := &messagingv1beta1.AuthorityRecord{}
		Expect(conversionK8sClient.Get(ctx, client.ObjectKeyFromObject(alpha), beta)).To(Succeed())
		Expect(beta.Spec.Authorities).To(ConsistOf("GET", "PUT"))
	})

	It("round-trips QueueManagerConnection v1alpha1 through v1beta1", func() {
		ctx := context.Background()
		alpha := &messagingv1alpha1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: "rt-qmc", Namespace: ns},
			Spec: messagingv1alpha1.QueueManagerConnectionSpec{
				QueueManager: "QM1",
				Endpoint:     "https://mq.example:9443",
				RESTPrefix:   "/ibmmq/rest/v3",
				CredentialsSecretRef: messagingv1alpha1.SecretReference{
					Name: "creds",
				},
				TLS: &messagingv1alpha1.TLSConfig{
					InsecureSkipVerify: true,
					CASecretRef:        &messagingv1alpha1.SecretReference{Name: "ca"},
				},
			},
			Status: messagingv1alpha1.QueueManagerConnectionStatus{},
		}
		alpha.Annotations = map[string]string{
			messagingv1alpha1.AllowInsecureTLSAnnotation: "true",
		}
		Expect(conversionK8sClient.Create(ctx, alpha)).To(Succeed())
		alpha.Status.ObservedGeneration = 1
		Expect(conversionK8sClient.Status().Update(ctx, alpha)).To(Succeed())

		beta := &messagingv1beta1.QueueManagerConnection{}
		Expect(conversionK8sClient.Get(ctx, client.ObjectKeyFromObject(alpha), beta)).To(Succeed())
		Expect(beta.Spec.TLS).NotTo(BeNil())
		Expect(beta.Spec.TLS.InsecureSkipVerify).To(BeTrue())
		Expect(beta.Spec.TLS.CASecretRef.Name).To(Equal("ca"))
		Expect(beta.Status.ObservedGeneration).To(Equal(int64(1)))
	})
})

func buildCRDInstallDir(root string) (string, error) {
	dir, err := os.MkdirTemp("", "mkurator-crd-install-*")
	if err != nil {
		return "", err
	}
	out := filepath.Join(dir, "crds.yaml")
	//nolint:gosec // test fixture path is constructed from repo root.
	cmd := exec.Command("go", "tool", "kustomize", "build", filepath.Join(root, "config", "crd"))
	cmd.Dir = root
	bundle, err := cmd.Output()
	if err != nil {
		return "", err
	}
	//nolint:gosec // envtest fixture permissions are not security-sensitive.
	if err := os.WriteFile(out, bundle, 0o644); err != nil {
		return "", err
	}
	return dir, nil
}

func conversionFixture(ctx context.Context, ns string) error {
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: ns}}
	if err := conversionK8sClient.Create(ctx, secret); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	caSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "ca", Namespace: ns}}
	if err := conversionK8sClient.Create(ctx, caSecret); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{
				Name: "creds",
			},
		},
	}
	if err := conversionK8sClient.Create(ctx, conn); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	ch := &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: "orders-app", Namespace: ns},
		Spec: messagingv1alpha1.ChannelSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
		},
	}
	if err := conversionK8sClient.Create(ctx, ch); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
