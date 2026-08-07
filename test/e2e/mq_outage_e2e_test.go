//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	mqOutageConnectionName = "e2e-qm1-outage"
	mqOutageMaxDepthV1     = "1000"
	mqOutageMaxDepthV2     = "2000"
)

// REQ-REL-2026-08 AC2, independent-review finding R2: "Given mqweb is unreachable for longer than
// one reconcile deadline, when it recovers, then every affected CR returns to Synced=True
// Available within one retry interval, and mqObjectExists=true matches the real MQ state."
//
// That acceptance criterion was only ever proven by unit and envtest coverage of the retry
// classification; nothing exercised a real outage against a real queue manager. This spec closes
// that gap end to end: it takes the kind platform's IBM MQ workload away, watches a Queue that was
// previously synced fall to Synced=False, brings mqweb back, and then asserts recovery happens
// with no further human action.
//
// Two deliberate deviations from the lane brief, both grounded in the code:
//
//   - The brief expected the QueueManagerConnection to go not-Ready during the outage. It does
//     not, by design: QueueManagerConnectionReconciler.fail() returns early (keeping Ready=True and
//     requeueing) when the connection is already Ready at the observed generation and the error is
//     transient, which suppresses flapping across a short mqweb blip. This spec therefore asserts
//     QMC readiness only after recovery.
//   - The brief expected the TerminalRetryInterval backstop (2m) to be the self-heal path. A dial
//     failure against an absent mqweb is wrapped by mqrest roundTrip as mqadmin.TransientError, so
//     the path actually under test is TransientRequeueInterval (30s) via setSyncedError — which is
//     exactly the path REQ-REL-2026-08 fixed. The spec pins that by asserting the failing
//     condition reason is the retryable "Error" and never "TerminalError": a regression to the
//     old wedge behaviour would surface here as a no-requeue terminal state.
var _ = Describe("Post-manager IBM MQ outage recovery", Label("mq", "mq-outage"), Serial, func() {
	var (
		ns          string
		prefix      string
		queueObject string
		queueCR     string
		statefulSet string
	)

	BeforeEach(func() {
		if !mqE2EEnabled() {
			Skip("IBM MQ e2e disabled; set KURATOR_E2E_MQ=1 and run task cluster:up")
		}
		e2eStage("MQ OUTAGE — mqweb outage and unattended recovery")
		ns = namespaceOutage
		ensureE2ENamespace(ns)
		waitForControllerAndWebhookReadyCached()
		ensureMQCredentialsSecret(ns)

		prefix = mqObjectPrefix()
		queueObject = mqOutageQueueObjectName(prefix)
		queueCR = mqCRName("e2e-outage-orders", prefix)
		statefulSet = mqPlatformStatefulSet()
	})

	// Serial: the whole cluster shares one IBM MQ queue manager, so taking it down would fail
	// every concurrent MQ spec. Not labelled "slow" on purpose — this is the regression guard for
	// a P0 production wedge (a CR stuck at Synced=False for 23h) and is worth the few minutes it
	// costs on every pull request, where the label filter is "(smoke || mq) && !slow".
	It("returns a Queue to Synced=True and the connection to Ready after an mqweb outage, unattended", func() {
		DeferCleanup(func() {
			if CurrentSpecReport().Failed() {
				dumpOutageDiagnostics(ns, queueCR, statefulSet)
			}
			// Restore mqweb before deleting the CRs: the operator's finalizers delete the MQ
			// objects over mqweb, and this also leaves the shared queue manager healthy for
			// whatever runs next.
			restoreMQWeb(statefulSet)
			kubectlForceRemoveNamespaced("queue", queueCR, ns)
			kubectlForceRemoveNamespaced("queuemanagerconnection", mqOutageConnectionName, ns)
		})

		By("establishing a healthy baseline: connection Ready and Queue synced on the queue manager")
		Expect(applyWithWebhookRetry(outageConnectionManifest(ns))).To(Succeed())
		eventuallyExpectOutageQMCReady(ns)

		Expect(applyWithWebhookRetry(outageQueueManifest(ns, queueCR, queueObject, mqOutageMaxDepthV1))).
			To(Succeed())
		eventuallyExpectQueueSynced(ns, queueCR)
		expectQueueMaxDepthOnMQ(queueObject, mqOutageMaxDepthV1)

		induceMQWebOutage(statefulSet)

		// Force a reconcile *while mqweb is down* by changing the desired maxdepth. This edit is
		// the trigger for the failing reconcile, never part of the recovery: nothing touches the
		// CR from here on. The passive alternative — waiting for the periodic drift resync — was
		// rejected because DriftResyncAfter jitters between 5 and 10 minutes, which would push the
		// spec past the suite's most generous Eventually helper for no extra coverage.
		By("applying a Queue spec change while mqweb is unreachable")
		Expect(applyWithWebhookRetry(outageQueueManifest(ns, queueCR, queueObject, mqOutageMaxDepthV2))).
			To(Succeed())

		By("expecting Synced=False with the retryable reason Error (never TerminalError)")
		Eventually(func(g Gomega) {
			status, reason, message := queueSyncedCondition(g, ns, queueCR)
			g.Expect(status).To(Equal("False"),
				"Queue %s should report Synced=False while mqweb is down (reason=%s message=%s)",
				queueCR, reason, message)
			g.Expect(reason).To(Equal("Error"),
				"a transient mqweb outage must stay retryable; reason=%s message=%s", reason, message)
		}).WithTimeout(mqSyncedEventuallyTimeout).WithPolling(5 * time.Second).Should(Succeed())

		generationBeforeRecovery := queueGeneration(ns, queueCR)

		restoreMQWeb(statefulSet)

		By("expecting the Queue to self-heal to Synced=True with no CR edit, annotation, or restart")
		Eventually(func(g Gomega) {
			status, reason, message := queueSyncedCondition(g, ns, queueCR)
			g.Expect(status).To(Equal("True"),
				"Queue %s should recover to Synced=True once mqweb is back (reason=%s message=%s)",
				queueCR, reason, message)
			g.Expect(reason).To(Equal("Available"), "recovered Queue should report Available, got %s", reason)

			exists, err := runKubectl("get", "queue", queueCR, "-n", ns,
				"-o", "jsonpath={.status.mqObjectExists}")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(strings.TrimSpace(exists)).To(Equal("true"),
				"Queue %s should observe the MQ object as existing after recovery", queueCR)
		}).WithTimeout(mqSyncedEventuallyTimeout).WithPolling(5 * time.Second).Should(Succeed())

		By("confirming recovery was unattended: the Queue spec was never touched after the outage")
		Expect(queueGeneration(ns, queueCR)).To(Equal(generationBeforeRecovery),
			"Queue %s generation changed during recovery, so the spec was edited, not self-healed", queueCR)

		By("confirming the connection recovered on its own")
		eventuallyExpectOutageQMCReady(ns)

		By("confirming mqObjectExists=true matches the real queue manager state")
		expectQueueMaxDepthOnMQ(queueObject, mqOutageMaxDepthV2)
	})
})

// queueSyncedCondition returns the Synced condition status, reason, and message of a Queue CR.
func queueSyncedCondition(g Gomega, ns, name string) (status, reason, message string) {
	out, err := runKubectl("get", "queue", name, "-n", ns, "-o",
		`jsonpath={.status.conditions[?(@.type=="Synced")].status}|`+
			`{.status.conditions[?(@.type=="Synced")].reason}|`+
			`{.status.conditions[?(@.type=="Synced")].message}`)
	g.Expect(err).NotTo(HaveOccurred())
	// The message is the last field, so a "|" inside it cannot corrupt status or reason.
	fields := strings.SplitN(out, "|", 3)
	for len(fields) < 3 {
		fields = append(fields, "")
	}
	return strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1]), strings.TrimSpace(fields[2])
}

// queueGeneration reads metadata.generation, which changes only when the spec is edited.
func queueGeneration(ns, name string) string {
	out, err := runKubectl("get", "queue", name, "-n", ns, "-o", "jsonpath={.metadata.generation}")
	Expect(err).NotTo(HaveOccurred(), "failed to read generation of Queue %s", name)
	return strings.TrimSpace(out)
}

func eventuallyExpectQueueSynced(ns, name string) {
	Eventually(func(g Gomega) {
		status, reason, message := queueSyncedCondition(g, ns, name)
		g.Expect(status).To(Equal("True"),
			"Queue %s should reach Synced=True (reason=%s message=%s)", name, reason, message)
	}).WithTimeout(mqSyncedEventuallyTimeout).WithPolling(5 * time.Second).Should(Succeed())
}

func eventuallyExpectOutageQMCReady(ns string) {
	Eventually(func(g Gomega) {
		out, err := runKubectl("get", "queuemanagerconnection", mqOutageConnectionName, "-n", ns,
			"-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}|`+
				`{.status.conditions[?(@.type=="Ready")].message}`)
		g.Expect(err).NotTo(HaveOccurred())
		fields := strings.SplitN(out, "|", 2)
		g.Expect(strings.TrimSpace(fields[0])).To(Equal("True"),
			"QueueManagerConnection %s should be Ready (message=%s)", mqOutageConnectionName, out)
	}).WithTimeout(qmcRotationEventuallyTimeout).WithPolling(5 * time.Second).Should(Succeed())
}

// expectQueueMaxDepthOnMQ asserts the real MQ state, so a status condition cannot pass on its own.
func expectQueueMaxDepthOnMQ(queueObject, maxDepth string) {
	client, err := newMQClient()
	Expect(err).NotTo(HaveOccurred())
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	state, err := client.GetQueue(ctx, e2eLocalQueueSpec(queueObject))
	Expect(err).NotTo(HaveOccurred(), "queue %s should exist on the queue manager", queueObject)
	Expect(state.Attributes["maxdepth"]).To(Equal(maxDepth),
		"queue %s should carry the desired maxdepth", queueObject)
}

// dumpOutageDiagnostics prints the state a failing CI run needs, since this spec cannot be
// reproduced without a kind cluster and a live queue manager.
func dumpOutageDiagnostics(ns, queueCR, statefulSet string) {
	if cond, err := runKubectl("get", "queue", queueCR, "-n", ns, "-o",
		`jsonpath=Synced={.status.conditions[?(@.type=="Synced")].status} `+
			`reason={.status.conditions[?(@.type=="Synced")].reason} `+
			`message={.status.conditions[?(@.type=="Synced")].message} `+
			`mqObjectExists={.status.mqObjectExists}`); err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "[diag] final Queue condition: %s\n", cond)
	}
	if cond, err := runKubectl("get", "queuemanagerconnection", mqOutageConnectionName, "-n", ns, "-o",
		`jsonpath=Ready={.status.conditions[?(@.type=="Ready")].status} `+
			`reason={.status.conditions[?(@.type=="Ready")].reason} `+
			`message={.status.conditions[?(@.type=="Ready")].message}`); err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "[diag] final QMC condition: %s\n", cond)
	}
	if sts, err := runKubectl("get", "statefulset", statefulSet, "-n", mqPlatformNamespace()); err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "[diag] IBM MQ StatefulSet:\n%s\n", sts)
	}
	if pods, err := runKubectl("get", "pods", "-n", mqPlatformNamespace()); err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "[diag] IBM MQ pods:\n%s\n", pods)
	}
	if logs, err := runKubectl("logs", "-n", namespace,
		"-l", "control-plane=controller-manager", "--tail=150", "--since=15m"); err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "[diag] controller-manager logs (tail 150):\n%s\n", logs)
	}
}

func outageConnectionManifest(ns string) string {
	return fmt.Sprintf(`apiVersion: messaging.mkurator.dev/v1beta1
kind: QueueManagerConnection
metadata:
  name: %s
  namespace: %s
  annotations:
    messaging.mkurator.dev/allow-insecure-tls: "true"
spec:
  queueManager: %s
  endpoint: https://ibm-mq.ibm-mq.svc:9443
  tls:
    insecureSkipVerify: true
  credentialsSecretRef:
    name: mq-credentials
`, mqOutageConnectionName, ns, envOr("KURATOR_E2E_MQ_QMGR", "QM1"))
}

func outageQueueManifest(ns, queueCR, queueObject, maxDepth string) string {
	return fmt.Sprintf(`apiVersion: messaging.mkurator.dev/v1beta1
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
    descr: e2e outage recovery queue
`, queueCR, ns, mqOutageConnectionName, queueObject, maxDepth)
}
