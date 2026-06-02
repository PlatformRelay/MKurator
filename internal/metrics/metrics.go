// Package metrics registers Kurator Prometheus collectors on controller-runtime's registry.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	ctrmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	ResultSuccess = "success"
	ResultError   = "error"
)

// Controller names used as the controller label value.
const (
	ControllerQueue                   = "queue"
	ControllerQueueManagerConnection  = "queuemanagerconnection"
)

// MQ operation names for mqweb adapter metrics.
const (
	MQOpPing        = "ping"
	MQOpGetQueue    = "get_queue"
	MQOpDefineQueue = "define_queue"
	MQOpDeleteQueue = "delete_queue"
	MQOpRunMQSC     = "run_mqsc"
)

var (
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kurator_reconcile_total",
			Help: "Total reconciliations by controller and result.",
		},
		[]string{"controller", "result"},
	)

	ReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kurator_reconcile_errors_total",
			Help: "Total reconcile passes that returned an error to the manager.",
		},
		[]string{"controller"},
	)

	MQOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kurator_mq_operations_total",
			Help: "Total mqweb operations by operation and result.",
		},
		[]string{"operation", "result"},
	)
)

func init() {
	ctrmetrics.Registry.MustRegister(
		ReconcileTotal,
		ReconcileErrors,
		MQOperationsTotal,
	)
}

// RecordReconcile increments reconcile counters for a controller pass.
func RecordReconcile(controller string, err error) {
	result := ResultSuccess
	if err != nil {
		result = ResultError
		ReconcileErrors.WithLabelValues(controller).Inc()
	}
	ReconcileTotal.WithLabelValues(controller, result).Inc()
}

// RecordMQOperation increments mqweb adapter operation counters.
func RecordMQOperation(operation string, err error) {
	result := ResultSuccess
	if err != nil {
		result = ResultError
	}
	MQOperationsTotal.WithLabelValues(operation, result).Inc()
}
