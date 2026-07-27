//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
)

var _ = ReportBeforeEach(func(report SpecReport) {
	e2eSpecLine("SPEC START", report.FullText())
})

var _ = ReportAfterEach(func(report SpecReport) {
	switch report.State {
	case types.SpecStatePassed:
		e2eSpecLine("SPEC PASS", report.FullText())
	case types.SpecStateFailed, types.SpecStatePanicked, types.SpecStateTimedout, types.SpecStateInterrupted:
		e2eSpecLine("SPEC FAIL", report.FullText())
		invalidateWebhookReadyCache()
		// Emit controller-manager logs to stdout on any failure so the CI log is
		// self-diagnosing without a cluster-level log capture step in the workflow.
		// Written to os.Stdout (not GinkgoWriter) so it always appears in raw CI output.
		if logs, err := runKubectl("logs", "-n", namespace,
			"-l", "control-plane=controller-manager",
			"--tail=100", "--since=15m"); err == nil {
			_, _ = fmt.Fprintf(os.Stdout, "[e2e-diag] controller-manager logs after spec failure %q:\n%s\n",
				report.FullText(), logs)
		}
	case types.SpecStateSkipped, types.SpecStatePending:
		e2eSpecLine("SPEC SKIP", report.FullText())
	default:
		e2eSpecLine("SPEC END", report.FullText())
	}
})
