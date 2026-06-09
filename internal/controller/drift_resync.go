package controller

import (
	"math/rand"
	"sync"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	defaultDriftResyncLower = 5 * time.Minute
	defaultDriftResyncUpper = 10 * time.Minute
)

var (
	driftResyncMu    sync.RWMutex
	driftResyncLower = defaultDriftResyncLower
	driftResyncUpper = defaultDriftResyncUpper
	//nolint:gosec // G404: math/rand is intentional for non-cryptographic requeue jitter.
	driftResyncRand = func() float64 { return rand.Float64() }
)

// SetDriftResyncInterval configures the jittered periodic resync window for synced workload CRs.
// lower must be positive and must not exceed upper; invalid values are ignored.
func SetDriftResyncInterval(lower, upper time.Duration) {
	if lower <= 0 || upper <= 0 || lower > upper {
		return
	}
	driftResyncMu.Lock()
	driftResyncLower = lower
	driftResyncUpper = upper
	driftResyncMu.Unlock()
}

// SetDriftResyncRand overrides the jitter source (tests only; nil restores the default).
func SetDriftResyncRand(fn func() float64) {
	driftResyncMu.Lock()
	if fn == nil {
		//nolint:gosec // G404: math/rand is intentional for non-cryptographic requeue jitter.
		driftResyncRand = func() float64 { return rand.Float64() }
	} else {
		driftResyncRand = fn
	}
	driftResyncMu.Unlock()
}

// DriftResyncLower returns the configured lower bound of the resync interval window.
func DriftResyncLower() time.Duration {
	driftResyncMu.RLock()
	defer driftResyncMu.RUnlock()
	return driftResyncLower
}

// DriftResyncUpper returns the configured upper bound of the resync interval window.
func DriftResyncUpper() time.Duration {
	driftResyncMu.RLock()
	defer driftResyncMu.RUnlock()
	return driftResyncUpper
}

// DriftResyncAfter returns a jittered duration in [lower, upper] for periodic drift resync.
func DriftResyncAfter() time.Duration {
	driftResyncMu.RLock()
	lower, upper := driftResyncLower, driftResyncUpper
	randFn := driftResyncRand
	driftResyncMu.RUnlock()

	if upper <= lower {
		return lower
	}
	jitter := randFn() * float64(upper-lower)
	return lower + time.Duration(jitter)
}

// workloadDriftResyncResult schedules the next drift check for a successfully synced workload CR.
func workloadDriftResyncResult() ctrl.Result {
	return ctrl.Result{RequeueAfter: DriftResyncAfter()}
}
