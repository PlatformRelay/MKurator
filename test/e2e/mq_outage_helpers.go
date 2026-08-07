//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// mqOutageProbeTimeout bounds a single reachability probe against mqweb. It is deliberately
// short: an unreachable mqweb fails on the first dial (connection refused / no route), and a
// healthy mqweb answers GET /admin/qmgr/<qm> well inside 20s even on a cold Liberty start.
const mqOutageProbeTimeout = 20 * time.Second

// mqPlatformNamespace is the namespace holding the kind platform's IBM MQ release
// (hack/kind-cluster/terraform/ibm-mq.tf).
func mqPlatformNamespace() string { return envOr("KURATOR_E2E_MQ_NAMESPACE", "ibm-mq") }

// mqPlatformStatefulSet resolves the IBM MQ StatefulSet name. The upstream chart derives it from
// the Helm release name, so it is discovered rather than hard-coded; KURATOR_E2E_MQ_STATEFULSET
// overrides discovery for clusters that host more than one StatefulSet in that namespace.
func mqPlatformStatefulSet() string {
	if name := os.Getenv("KURATOR_E2E_MQ_STATEFULSET"); name != "" {
		return name
	}
	out, err := runKubectl("get", "statefulset", "-n", mqPlatformNamespace(),
		"-o", "jsonpath={.items[0].metadata.name}")
	Expect(err).NotTo(HaveOccurred(), "failed to list StatefulSets in namespace %s", mqPlatformNamespace())
	name := strings.TrimSpace(out)
	Expect(name).NotTo(BeEmpty(),
		"no IBM MQ StatefulSet found in namespace %s; set KURATOR_E2E_MQ_STATEFULSET", mqPlatformNamespace())
	return name
}

// scaleMQPlatform scales the IBM MQ StatefulSet. The queue manager keeps its PVC, so scaling back
// up restores the same queue manager and every MQ object created by earlier specs.
func scaleMQPlatform(statefulSet string, replicas int) {
	_, err := runKubectl("scale", "statefulset", statefulSet, "-n", mqPlatformNamespace(),
		fmt.Sprintf("--replicas=%d", replicas))
	Expect(err).NotTo(HaveOccurred(),
		"failed to scale StatefulSet %s/%s to %d", mqPlatformNamespace(), statefulSet, replicas)
}

// mqWebReachable reports whether a freshly built mqweb client can complete an authenticated
// request. A new client per probe guarantees the answer reflects the current state of mqweb and
// never a pooled keep-alive connection.
func mqWebReachable() bool {
	client, err := newMQClient()
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), mqOutageProbeTimeout)
	defer cancel()
	return client.Ping(ctx) == nil
}

// induceMQWebOutage takes mqweb down by scaling the IBM MQ workload to zero, and blocks until the
// suite can confirm mqweb no longer answers.
//
// Scaling the workload away is the only mechanism available to the suite that produces a genuine
// outage. Breaking the Service selector or installing a NetworkPolicy removes future routing but
// leaves already-established keep-alive connections from the operator usable, so a reconcile that
// reuses an idle connection would still succeed and the spec would be proving nothing. Removing
// the pod severs every connection and takes the listener away.
func induceMQWebOutage(statefulSet string) {
	By("scaling the IBM MQ workload to zero so mqweb becomes unreachable")
	scaleMQPlatform(statefulSet, 0)

	Eventually(func(g Gomega) {
		out, err := runKubectl("get", "statefulset", statefulSet, "-n", mqPlatformNamespace(),
			"-o", "jsonpath={.status.replicas}")
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(strings.TrimSpace(out)).To(BeElementOf("", "0"),
			"IBM MQ StatefulSet %s should have no pods left", statefulSet)
	}).WithTimeout(KubectlWaitDuration).WithPolling(2 * time.Second).Should(Succeed())

	By("confirming mqweb no longer answers")
	Eventually(func(g Gomega) {
		g.Expect(mqWebReachable()).To(BeFalse(), "mqweb should be unreachable during the induced outage")
	}).WithTimeout(KubectlWaitDuration).WithPolling(3 * time.Second).Should(Succeed())
}

// restoreMQWeb scales the IBM MQ workload back up and blocks until mqweb answers again. It is
// idempotent, so it doubles as the DeferCleanup guard when a spec fails mid-outage.
//
// This is the ONLY action a recovery assertion may take: no custom resource is edited, annotated,
// re-applied, or deleted afterwards, and the operator is not restarted.
func restoreMQWeb(statefulSet string) {
	By("scaling the IBM MQ workload back up")
	scaleMQPlatform(statefulSet, 1)

	Eventually(func(g Gomega) {
		out, err := runKubectl("get", "statefulset", statefulSet, "-n", mqPlatformNamespace(),
			"-o", "jsonpath={.status.readyReplicas}")
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(strings.TrimSpace(out)).To(Equal("1"),
			"IBM MQ StatefulSet %s should report a ready replica again", statefulSet)
	}).WithTimeout(qmcWatchRecoveryEventuallyTimeout).WithPolling(5 * time.Second).Should(Succeed())

	By("waiting for mqweb to answer again")
	Eventually(func(g Gomega) {
		g.Expect(mqWebReachable()).To(BeTrue(), "mqweb should answer once the queue manager is back")
	}).WithTimeout(qmcWatchRecoveryEventuallyTimeout).WithPolling(5 * time.Second).Should(Succeed())
}
