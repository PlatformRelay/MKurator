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

// qmcWatchRecoveryEventuallyTimeout covers the watch-driven + backstop recovery path
// (AUTH-14): a union auth-Secret is rotated with NO QMC spec change / re-apply. The
// fast path is the Secret-watch enqueue → reconcile → client rebuild → Ping. The
// 2-minute backstop requeue (TerminalRetryInterval) guards against silent enqueue drops
// (e.g. transient hub re-read failure under kind/CI load). Worst-case recovery is:
// 2m (backstop fires) + reconcile overhead ≈ 2.5m; 5m gives ample headroom.
const qmcWatchRecoveryEventuallyTimeout = 5 * time.Minute
