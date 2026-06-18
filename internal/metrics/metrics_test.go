package metrics

import (
	"errors"
	"fmt"
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

func TestRecordReconcile_errorIncrementsResultErrorTotal(t *testing.T) {
	t.Parallel()
	before := counterValue(t, ReconcileTotal, ControllerTopic, ResultError)
	RecordReconcile(ControllerTopic, errors.New("reconcile failed"))
	after := counterValue(t, ReconcileTotal, ControllerTopic, ResultError)
	if after != before+1 {
		t.Fatalf("error result total: before=%v after=%v", before, after)
	}
}

func TestRecordReconcile_successDoesNotIncrementErrors(t *testing.T) {
	t.Parallel()
	before := counterValue(t, ReconcileErrors, ControllerChannel)
	RecordReconcile(ControllerChannel, nil)
	after := counterValue(t, ReconcileErrors, ControllerChannel)
	if after != before {
		t.Fatalf("errors should be unchanged: before=%v after=%v", before, after)
	}
}

func TestRecordReconcile_wrappedErrorCountsAsError(t *testing.T) {
	t.Parallel()
	root := errors.New("mqweb timeout")
	wrapped := fmt.Errorf("define queue: %w", root)

	errBefore := counterValue(t, ReconcileErrors, ControllerQueueManagerConnection)
	totalBefore := counterValue(t, ReconcileTotal, ControllerQueueManagerConnection, ResultError)
	RecordReconcile(ControllerQueueManagerConnection, wrapped)
	errAfter := counterValue(t, ReconcileErrors, ControllerQueueManagerConnection)
	totalAfter := counterValue(t, ReconcileTotal, ControllerQueueManagerConnection, ResultError)
	if errAfter != errBefore+1 {
		t.Fatalf("wrapped error count: before=%v after=%v", errBefore, errAfter)
	}
	if totalAfter != totalBefore+1 {
		t.Fatalf("wrapped error total: before=%v after=%v", totalBefore, totalAfter)
	}
}

func TestRecordReconcile_controllerLabelCardinality(t *testing.T) {
	t.Parallel()
	controllers := []string{
		ControllerQueue,
		ControllerTopic,
		ControllerChannel,
		ControllerChannelAuthRule,
		ControllerAuthorityRecord,
		ControllerQueueManagerConnection,
	}
	for _, ctrl := range controllers {
		t.Run(ctrl, func(t *testing.T) {
			t.Parallel()
			before := counterValue(t, ReconcileTotal, ctrl, ResultSuccess)
			RecordReconcile(ctrl, nil)
			after := counterValue(t, ReconcileTotal, ctrl, ResultSuccess)
			if after != before+1 {
				t.Fatalf("controller %q: before=%v after=%v", ctrl, before, after)
			}
		})
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

func TestRecordDriftDetected_controllerLabelCardinality(t *testing.T) {
	t.Parallel()
	workloadControllers := []string{
		ControllerQueue,
		ControllerTopic,
		ControllerChannel,
		ControllerChannelAuthRule,
		ControllerAuthorityRecord,
	}
	for _, ctrl := range workloadControllers {
		t.Run(ctrl, func(t *testing.T) {
			t.Parallel()
			before := counterValue(t, DriftDetectedTotal, ctrl)
			RecordDriftDetected(ctrl)
			after := counterValue(t, DriftDetectedTotal, ctrl)
			if after != before+1 {
				t.Fatalf("controller %q: before=%v after=%v", ctrl, before, after)
			}
		})
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

func TestRecordMQOperation_successDoesNotIncrementErrorResult(t *testing.T) {
	t.Parallel()
	before := counterValue(t, MQOperationsTotal, MQOpGetQueue, ResultError)
	RecordMQOperation(MQOpGetQueue, nil)
	after := counterValue(t, MQOperationsTotal, MQOpGetQueue, ResultError)
	if after != before {
		t.Fatalf("error result should be unchanged: before=%v after=%v", before, after)
	}
}

func TestRecordMQOperation_wrappedErrorCountsAsError(t *testing.T) {
	t.Parallel()
	wrapped := fmt.Errorf("get topic: %w", errors.New("404"))

	before := counterValue(t, MQOperationsTotal, MQOpGetTopic, ResultError)
	RecordMQOperation(MQOpGetTopic, wrapped)
	after := counterValue(t, MQOperationsTotal, MQOpGetTopic, ResultError)
	if after != before+1 {
		t.Fatalf("wrapped error count: before=%v after=%v", before, after)
	}
}

func TestRecordMQOperation_operationLabelCardinality(t *testing.T) {
	t.Parallel()
	operations := []string{
		MQOpPing,
		MQOpGetQueue,
		MQOpDefineQueue,
		MQOpDeleteQueue,
		MQOpGetTopic,
		MQOpDefineTopic,
		MQOpDeleteTopic,
		MQOpGetChannel,
		MQOpDefineChannel,
		MQOpDeleteChannel,
		MQOpSetChannelAuth,
		MQOpGetChannelAuth,
		MQOpDeleteChannelAuth,
		MQOpSetAuthority,
		MQOpGetAuthority,
		MQOpDeleteAuthority,
		MQOpRunMQSC,
	}
	for _, op := range operations {
		t.Run(op, func(t *testing.T) {
			t.Parallel()
			before := counterValue(t, MQOperationsTotal, op, ResultSuccess)
			RecordMQOperation(op, nil)
			after := counterValue(t, MQOperationsTotal, op, ResultSuccess)
			if after != before+1 {
				t.Fatalf("operation %q: before=%v after=%v", op, before, after)
			}
		})
	}
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

func TestRecordCircuitBreakerTransition_labelCardinality(t *testing.T) {
	t.Parallel()
	transitions := [][2]string{
		{"closed", "open"},
		{"open", "half_open"},
		{"half_open", "closed"},
		{"half_open", "open"},
	}
	for _, pair := range transitions {
		from, to := pair[0], pair[1]
		t.Run(from+"->"+to, func(t *testing.T) {
			t.Parallel()
			before := counterValue(t, CircuitBreakerTransitionsTotal, from, to)
			RecordCircuitBreakerTransition(from, to)
			after := counterValue(t, CircuitBreakerTransitionsTotal, from, to)
			if after != before+1 {
				t.Fatalf("transition %s->%s: before=%v after=%v", from, to, before, after)
			}
		})
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
