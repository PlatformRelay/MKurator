package controller

import (
	"context"
	"sync"
	"time"
)

const defaultMQRequestTimeout = 30 * time.Second

var (
	mqRequestTimeoutMu sync.RWMutex
	mqRequestTimeout   = defaultMQRequestTimeout
)

func SetMQRequestTimeout(d time.Duration) {
	if d <= 0 {
		return
	}
	mqRequestTimeoutMu.Lock()
	mqRequestTimeout = d
	mqRequestTimeoutMu.Unlock()
}

func MQRequestTimeout() time.Duration {
	mqRequestTimeoutMu.RLock()
	defer mqRequestTimeoutMu.RUnlock()
	return mqRequestTimeout
}

func MQRequestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, MQRequestTimeout())
}
