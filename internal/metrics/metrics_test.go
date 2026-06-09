package metrics

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordReconcile(t *testing.T) {
	t.Parallel()
	before := counterValue(t, ReconcileTotal, ControllerQueue, ResultSuccess)
	RecordReconcile(ControllerQueue, nil)
	after := counterValue(t, ReconcileTotal, ControllerQueue, ResultSuccess)
	if after != before+1 {
		t.Fatalf("success count: before=%v after=%v", before, after)
	}

	errBefore := counterValue(t, ReconcileErrors, ControllerQueue)
	RecordReconcile(ControllerQueue, errors.New("boom"))
	errAfter := counterValue(t, ReconcileErrors, ControllerQueue)
	if errAfter != errBefore+1 {
		t.Fatalf("error count: before=%v after=%v", errBefore, errAfter)
	}
}

func TestRecordDriftDetected(t *testing.T) {
	t.Parallel()
	before := counterValue(t, DriftDetectedTotal, ControllerQueue)
	RecordDriftDetected(ControllerQueue)
	after := counterValue(t, DriftDetectedTotal, ControllerQueue)
	if after != before+1 {
		t.Fatalf("drift count: before=%v after=%v", before, after)
	}
}

func TestRecordMQOperation(t *testing.T) {
	t.Parallel()
	before := counterValue(t, MQOperationsTotal, MQOpPing, ResultSuccess)
	RecordMQOperation(MQOpPing, nil)
	after := counterValue(t, MQOperationsTotal, MQOpPing, ResultSuccess)
	if after != before+1 {
		t.Fatalf("mq op count: before=%v after=%v", before, after)
	}

	errBefore := counterValue(t, MQOperationsTotal, MQOpDefineQueue, ResultError)
	RecordMQOperation(MQOpDefineQueue, errors.New("mqweb down"))
	errAfter := counterValue(t, MQOperationsTotal, MQOpDefineQueue, ResultError)
	if errAfter != errBefore+1 {
		t.Fatalf("mq op error count: before=%v after=%v", errBefore, errAfter)
	}
}

func counterValue(t *testing.T, cv *prometheus.CounterVec, labels ...string) float64 {
	t.Helper()
	m, err := cv.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatal(err)
	}
	return testutil.ToFloat64(m)
}

func TestRecordCircuitBreakerTransition(t *testing.T) {
	t.Parallel()
	before := counterValue(t, CircuitBreakerTransitionsTotal, "closed", "open")
	RecordCircuitBreakerTransition("closed", "open")
	after := counterValue(t, CircuitBreakerTransitionsTotal, "closed", "open")
	if after != before+1 {
		t.Fatalf("transition count: before=%v after=%v", before, after)
	}
}
