//go:build e2e
// +build e2e

package e2e

import "time"

// mqAuthrecCleanupEventuallyTimeout covers MQ-side AUTHREC removal after CR delete.
const mqAuthrecCleanupEventuallyTimeout = 8 * time.Minute

// mqSyncedEventuallyTimeout is the default Synced/Ready wait for MQ CR specs (not QMC rotation).
const mqSyncedEventuallyTimeout = 3 * time.Minute

// qmcRotationEventuallyTimeout covers secret rotation and QMC recreate paths.
const qmcRotationEventuallyTimeout = 3 * time.Minute

// qmcWatchRecoveryEventuallyTimeout covers the pure watch-driven recovery path
// (AUTH-14): a union auth-Secret is rotated with NO QMC spec change / re-apply, so
// readiness recovers solely via the Secret-watch enqueue → reconcile → client
// rebuild → live mqweb Ping. That chain has no periodic-resync backstop by design
// (ADR-0027/ADR-0023), so it is legitimately slower and more variable under kind/CI
// load than the re-apply paths — 3m proved marginal (flaked on identical code). A
// total watch miss still fails at 5m, so this loosens timing without masking a bug.
const qmcWatchRecoveryEventuallyTimeout = 5 * time.Minute
