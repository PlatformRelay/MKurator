//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/konih/kurator/test/utils"
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

	waitForControllerAndWebhookReady()
}

// waitForControllerAndWebhookReady blocks until cert-manager has issued the webhook
// TLS secret and the controller-manager pod reports Ready (webhook listener up).
func waitForControllerAndWebhookReady() {
	Eventually(func(g Gomega) {
		cmd := exec.Command("kubectl", "get", "secret", "webhook-server-cert", "-n", namespace)
		_, err := utils.Run(cmd)
		g.Expect(err).NotTo(HaveOccurred(), "webhook-server-cert should exist")
	}).WithTimeout(3 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())

	Eventually(func(g Gomega) {
		cmd := exec.Command("kubectl", "get", "pods", "-n", namespace,
			"-l", "control-plane=controller-manager",
			"-o", "jsonpath={.items[0].status.conditions[?(@.type=='Ready')].status}")
		out, err := utils.Run(cmd)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(out).To(Equal("True"), "controller-manager should be Ready")
	}).WithTimeout(5 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())

	Eventually(func(g Gomega) {
		cmd := exec.Command("kubectl", "get", "endpoints", "kurator-webhook-service", "-n", namespace,
			"-o", "jsonpath={.subsets[0].addresses[0].ip}")
		out, err := utils.Run(cmd)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(out).NotTo(BeEmpty(), "webhook service should have endpoints")
	}).WithTimeout(5 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())
}
