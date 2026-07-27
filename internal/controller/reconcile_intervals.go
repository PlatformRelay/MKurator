package controller

import (
	"sync"
	"time"
)

const (
	defaultConnectionWaitInterval   = 15 * time.Second
	defaultTransientRequeueInterval = 30 * time.Second
	// defaultTerminalRetryInterval is the backstop requeue delay after a terminal error
	// (e.g. 401 Unauthorized). A periodic retry ensures the controller self-heals if the
	// Secret-watch enqueue was dropped (e.g. transient hub re-read failure in a loaded cluster)
	// without relying solely on a future watch event — closing a silent-stuck gap (AUTH-14).
	defaultTerminalRetryInterval = 2 * time.Minute
)

var (
	reconcileIntervalsMu     sync.RWMutex
	connectionWaitInterval   = defaultConnectionWaitInterval
	transientRequeueInterval = defaultTransientRequeueInterval
	terminalRetryInterval    = defaultTerminalRetryInterval
)

// SetConnectionWaitInterval configures the requeue delay while waiting for a QueueManagerConnection.
// Non-positive values are ignored.
func SetConnectionWaitInterval(d time.Duration) {
	if d <= 0 {
		return
	}
	reconcileIntervalsMu.Lock()
	connectionWaitInterval = d
	reconcileIntervalsMu.Unlock()
}

// SetTransientRequeueInterval configures the requeue delay after transient MQ or connection errors.
// Non-positive values are ignored.
func SetTransientRequeueInterval(d time.Duration) {
	if d <= 0 {
		return
	}
	reconcileIntervalsMu.Lock()
	transientRequeueInterval = d
	reconcileIntervalsMu.Unlock()
}

// ConnectionWaitInterval returns the configured connection-wait requeue delay.
func ConnectionWaitInterval() time.Duration {
	reconcileIntervalsMu.RLock()
	defer reconcileIntervalsMu.RUnlock()
	return connectionWaitInterval
}

// TransientRequeueInterval returns the configured transient-error requeue delay.
func TransientRequeueInterval() time.Duration {
	reconcileIntervalsMu.RLock()
	defer reconcileIntervalsMu.RUnlock()
	return transientRequeueInterval
}

// SetTerminalRetryInterval configures the backstop requeue delay after terminal MQ errors.
// Non-positive values are ignored.
func SetTerminalRetryInterval(d time.Duration) {
	if d <= 0 {
		return
	}
	reconcileIntervalsMu.Lock()
	terminalRetryInterval = d
	reconcileIntervalsMu.Unlock()
}

// TerminalRetryInterval returns the configured terminal-error backstop requeue delay.
func TerminalRetryInterval() time.Duration {
	reconcileIntervalsMu.RLock()
	defer reconcileIntervalsMu.RUnlock()
	return terminalRetryInterval
}
