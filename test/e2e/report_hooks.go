//go:build e2e
// +build e2e

package e2e

import (
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
	case types.SpecStateSkipped, types.SpecStatePending:
		e2eSpecLine("SPEC SKIP", report.FullText())
	default:
		e2eSpecLine("SPEC END", report.FullText())
	}
})
