package controller

import (
	"sync"
	"time"
)

const (
	defaultConnectionWaitInterval   = 15 * time.Second
	defaultTransientRequeueInterval = 30 * time.Second
)

var (
	reconcileIntervalsMu     sync.RWMutex
	connectionWaitInterval   = defaultConnectionWaitInterval
	transientRequeueInterval = defaultTransientRequeueInterval
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
