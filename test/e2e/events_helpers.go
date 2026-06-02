//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
	"github.com/konih/kurator/test/utils"
)

// eventuallyExpectObjectEvent waits for a Kubernetes Event on the named CR.
func eventuallyExpectObjectEvent(ns, kind, name, eventType, reason string) {
	Eventually(func(g Gomega) {
		g.Expect(hasObjectEvent(ns, kind, name, eventType, reason)).To(BeTrue(),
			"expected %s event with reason %q on %s/%s", eventType, reason, kind, name)
	}).WithTimeout(30 * time.Second).WithPolling(2 * time.Second).Should(Succeed())
}

func hasObjectEvent(ns, kind, name, eventType, reason string) (bool, error) {
	cmd := exec.Command("kubectl", "get", "events", "-n", ns,
		"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s", name, kind),
		"-o", "jsonpath={range .items[*]}{.type}{\" \"}{.reason}{\"\\n\"}{end}",
	)
	out, err := utils.Run(cmd)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if fields[0] == eventType && fields[1] == reason {
			return true, nil
		}
	}
	return false, nil
}

func eventuallyExpectQueueAvailableEvent(ns, queueName string) {
	By("checking for Normal Available event on Queue")
	eventuallyExpectObjectEvent(ns, "Queue", queueName, "Normal", messagingv1alpha1.ReasonAvailable)
}
