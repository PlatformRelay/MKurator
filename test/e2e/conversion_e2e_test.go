//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	conversionQueueMaxDepth = "1000"
	conversionQueueDescr    = "e2e conversion orders queue"
)

// Serial: upgradeMKuratorCRDs re-applies the cluster-wide CRD bundle mid-spec;
// parallel processes applying CRs against those CRDs would race the upgrade.
var _ = Describe("v1alpha1 to v1beta1 CRD upgrade", Serial, Label("conversion", "mq"), func() {
	var (
		ns          string
		prefix      string
		queueObject string
		queueCR     string
	)

	BeforeEach(func() {
		if !mqE2EEnabled() {
			Skip("IBM MQ e2e disabled; set KURATOR_E2E_MQ=1 and run task cluster:up")
		}
		if !webhookReady.Load() {
			waitForControllerAndWebhookReadyCached()
		}

		ns = namespaceQueues
		ensureE2ENamespace(ns)
		prefix = mqObjectPrefix()
		queueObject = mqQueueObjectName(prefix)
		queueCR = mqCRName("e2e-conversion", prefix)
		ensureMQCredentialsSecret(ns)
		DeferCleanup(func() {
			cleanupMQSpec(ns, "queue", queueCR)
		})
	})

	It("converts stored v1alpha1 Queue to served v1beta1 after CRD upgrade", func() {
		Expect(applyWithWebhookRetry(connectionManifest(ns))).To(Succeed())
		eventuallyExpectQMCReady(ns)

		queueYAML := fmt.Sprintf(`apiVersion: messaging.mkurator.dev/v1alpha1
kind: Queue
metadata:
  name: %s
  namespace: %s
spec:
  connectionRef:
    name: %s
  queueName: %s
  type: local
  attributes:
    maxdepth: "%s"
    descr: %s
`, queueCR, ns, mqConnectionName, queueObject, conversionQueueMaxDepth, conversionQueueDescr)
		Expect(applyWithWebhookRetry(queueYAML)).To(Succeed())
		eventuallyExpectQueueSynced(ns, queueCR)

		By("re-applying upgraded CRD bundle with conversion webhook")
		upgradeMKuratorCRDs()

		By("verifying stored v1alpha1 representation still has attribute map")
		Eventually(func(g Gomega) {
			apiVersion, maxDepthAttr, descrAttr, err := queueStoredV1Alpha1Fields(ns, queueCR)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(apiVersion).To(Equal("messaging.mkurator.dev/v1alpha1"))
			g.Expect(maxDepthAttr).To(Equal(conversionQueueMaxDepth))
			g.Expect(descrAttr).To(Equal(conversionQueueDescr))
		}).WithTimeout(2 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())

		By("verifying served v1beta1 representation folds attributes into typed fields")
		Eventually(func(g Gomega) {
			apiVersion, maxDepth, description, attrsJSON, err := queueServedV1Beta1Fields(ns, queueCR)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(apiVersion).To(Equal("messaging.mkurator.dev/v1beta1"))
			g.Expect(maxDepth).To(Equal(conversionQueueMaxDepth))
			g.Expect(description).To(Equal(conversionQueueDescr))
			g.Expect(attrsJSON).To(Or(BeEmpty(), Equal("map[]"), Equal("{}")),
				"folded attribute keys should not remain in spec.attributes on v1beta1 read")
		}).WithTimeout(2 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())

		By("verifying reconcile stays green after CRD upgrade")
		eventuallyExpectQueueSynced(ns, queueCR)

		client, err := newMQClient()
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		state, err := client.GetQueue(ctx, e2eLocalQueueSpec(queueObject))
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Attributes["maxdepth"]).To(Equal(conversionQueueMaxDepth),
			"MQ queue depth should be unchanged after CRD upgrade and conversion")
	})
})

func eventuallyExpectQueueSynced(ns, queueCR string) {
	Eventually(func(g Gomega) {
		out, err := runKubectl("get", "queue", queueCR, "-n", ns,
			"-o", "jsonpath={.status.conditions[?(@.type==\"Synced\")].status}")
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(out).To(Equal("True"), "Queue %s should stay Synced", queueCR)
	}).WithTimeout(mqSyncedEventuallyTimeout).WithPolling(5 * time.Second).Should(Succeed())
}

func queueStoredV1Alpha1Fields(ns, name string) (apiVersion, maxDepth, descr string, err error) {
	apiVersion, err = runKubectl("get", "queues.v1alpha1.messaging.mkurator.dev", name, "-n", ns,
		"-o", "jsonpath={.apiVersion}")
	if err != nil {
		return "", "", "", err
	}
	maxDepth, err = runKubectl("get", "queues.v1alpha1.messaging.mkurator.dev", name, "-n", ns,
		"-o", "jsonpath={.spec.attributes.maxdepth}")
	if err != nil {
		return "", "", "", err
	}
	descr, err = runKubectl("get", "queues.v1alpha1.messaging.mkurator.dev", name, "-n", ns,
		"-o", "jsonpath={.spec.attributes.descr}")
	if err != nil {
		return "", "", "", err
	}
	return apiVersion, maxDepth, descr, nil
}

func queueServedV1Beta1Fields(ns, name string) (apiVersion, maxDepth, description, attrsJSON string, err error) {
	resource := "queues.v1beta1.messaging.mkurator.dev"
	apiVersion, err = runKubectl("get", resource, name, "-n", ns, "-o", "jsonpath={.apiVersion}")
	if err != nil {
		return "", "", "", "", err
	}
	maxDepth, err = runKubectl("get", resource, name, "-n", ns, "-o", "jsonpath={.spec.maxDepth}")
	if err != nil {
		return "", "", "", "", err
	}
	description, err = runKubectl("get", resource, name, "-n", ns, "-o", "jsonpath={.spec.description}")
	if err != nil {
		return "", "", "", "", err
	}
	attrsJSON, err = runKubectl("get", resource, name, "-n", ns, "-o", "jsonpath={.spec.attributes}")
	if err != nil {
		return "", "", "", "", err
	}
	return apiVersion, maxDepth, description, attrsJSON, nil
}
