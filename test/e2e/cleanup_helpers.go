//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"

	"github.com/konih/kurator/test/utils"
)

const e2eCleanupTimeout = 5 * time.Minute

// kuratorMQResources are namespaced messaging.kurator.dev kinds (kubectl resource names).
var kuratorMQResources = []string{
	"queuemanagerconnection",
	"queue",
	"topic",
	"channel",
	"channelauthrule",
	"authorityrecord",
}

func e2eCleanupNamespaces() []string {
	return append([]string{namespace}, mqE2ENamespaces...)
}

func skipCRDUndeploy() bool {
	return os.Getenv("KURATOR_E2E_SKIP_CRD_UNDEPLOY") == "1"
}

// cleanupE2EResources tears down MQ CRs, the operator, CRDs, and e2e namespaces.
// It is bounded by e2eCleanupTimeout so AfterSuite cannot hang indefinitely.
func cleanupE2EResources() {
	By(fmt.Sprintf("cleaning up e2e resources (max %s)", e2eCleanupTimeout))
	done := make(chan struct{})
	go func() {
		defer close(done)
		cleanupE2EResourcesOnce()
	}()
	select {
	case <-done:
	case <-time.After(e2eCleanupTimeout):
		_, _ = fmt.Fprintf(GinkgoWriter,
			"e2e cleanup timed out after %s (kill the test process if kubectl is still stuck)\n",
			e2eCleanupTimeout,
		)
	}
}

func cleanupE2EResourcesOnce() {
	namespaces := e2eCleanupNamespaces()
	undeployOperatorForE2E()
	deleteE2ENamespacesNoWait(namespaces)
}

// deleteAllE2ECustomResources removes messaging.kurator.dev CRs from kurator-e2e-* and
// kurator-system without waiting on operator finalizers, then strips stuck finalizers.
// Run before CRD delete so kubectl does not block on InstanceDeletionInProgress.
func deleteAllE2ECustomResources() {
	namespaces := e2eCleanupNamespaces()
	deleteAllKuratorCRsNoWait(namespaces)
	stripRemainingFinalizers(namespaces)
}

func deleteAllKuratorCRsNoWait(namespaces []string) {
	By("deleting Kurator custom resources in e2e namespaces (no wait)")
	for _, ns := range namespaces {
		for _, res := range kuratorMQResources {
			cmd := exec.Command("kubectl", "delete", res, "--all", "-n", ns,
				"--ignore-not-found", "--wait=false")
			_, _ = utils.Run(cmd)
		}
	}
}

// stripRemainingFinalizers removes operator finalizers when the controller is already gone.
func stripRemainingFinalizers(namespaces []string) {
	const patch = `{"metadata":{"finalizers":null}}`
	for _, ns := range namespaces {
		for _, res := range kuratorMQResources {
			cmd := exec.Command("kubectl", "get", res, "-n", ns,
				"-o", "jsonpath={range .items[*]}{.metadata.name}{\"\\n\"}{end}")
			out, err := utils.Run(cmd)
			if err != nil {
				continue
			}
			for _, name := range strings.Split(strings.TrimSpace(out), "\n") {
				if name == "" {
					continue
				}
				patchCmd := exec.Command("kubectl", "patch", res, name, "-n", ns,
					"--type=merge", "-p", patch)
				_, _ = utils.Run(patchCmd)
			}
		}
	}
}

func deleteE2ENamespacesNoWait(namespaces []string) {
	By("deleting e2e namespaces (no wait)")
	for _, ns := range namespaces {
		cmd := exec.Command("kubectl", "delete", "ns", ns, "--ignore-not-found", "--wait=false")
		_, _ = utils.Run(cmd)
	}
}

func undeployKustomizeOperatorNoWait() {
	By("removing controller-manager manifests (no wait)")
	cmd := exec.Command("kubectl", "delete", "--ignore-not-found", "-k", "config/default", "--wait=false")
	cmd.Env = taskEnv()
	_, _ = utils.Run(cmd)
}

func undeployKuratorCRDsNoWait() {
	if skipCRDUndeploy() {
		_, _ = fmt.Fprintf(GinkgoWriter,
			"Skipping Kurator CRD delete (KURATOR_E2E_SKIP_CRD_UNDEPLOY=1)\n")
		return
	}
	By("removing Kurator CRDs (no wait)")
	projectDir, err := utils.GetProjectDir()
	if err != nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "undeploy CRDs: get project dir: %v\n", err)
		return
	}
	crdDir := filepath.Join(projectDir, "config", "crd", "bases")
	cmd := exec.Command("kubectl", "delete", "--ignore-not-found", "-f", crdDir, "--wait=false")
	cmd.Env = taskEnv()
	_, _ = utils.Run(cmd)
}
