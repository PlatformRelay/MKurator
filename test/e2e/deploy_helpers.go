//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/konradheimel/kurator/test/utils"
)

// ensureOperatorForMQE2E (re)creates the operator install needed by Queue reconciliation tests.
// The Manager suite tears down the namespace in AfterAll; MQ specs run afterward.
func ensureOperatorForMQE2E() {
	By("creating manager namespace for MQ e2e")
	Expect(kubectlApply(fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, namespace))).To(Succeed())

	By("labeling the namespace to enforce the restricted security policy")
	cmd := exec.Command("kubectl", "label", "--overwrite", "ns", namespace,
		"pod-security.kubernetes.io/enforce=restricted")
	_, err := utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with restricted policy")

	By("installing CRDs for MQ e2e")
	cmd = exec.Command("make", "install")
	_, err = utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to install CRDs")

	Eventually(func(g Gomega) {
		check := exec.Command("kubectl", "get", "crd", "queuemanagerconnections.messaging.kurator.dev")
		_, runErr := utils.Run(check)
		g.Expect(runErr).NotTo(HaveOccurred(), "QueueManagerConnection CRD should be registered")
	}).WithTimeout(2 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())

	By("deploying the controller-manager for MQ e2e")
	cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", managerImage))
	_, err = utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to deploy the controller-manager")

	Eventually(func(g Gomega) {
		cmd := exec.Command("kubectl", "get", "pods", "-n", namespace,
			"-l", "control-plane=controller-manager",
			"-o", "jsonpath={.items[0].status.phase}")
		out, runErr := utils.Run(cmd)
		g.Expect(runErr).NotTo(HaveOccurred())
		g.Expect(out).To(Equal("Running"))
	}).WithTimeout(3 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())
}
