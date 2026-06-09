// Package metrics registers MKurator Prometheus collectors on controller-runtime's registry.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	ctrmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	ResultSuccess = "success"
	ResultError   = "error"

	labelController = "controller"
	labelResult     = "result"
	labelOperation  = "operation"
)

// Controller names used as the controller label value.
const (
	ControllerQueue                  = "queue"
	ControllerTopic                  = "topic"
	ControllerChannel                = "channel"
	ControllerChannelAuthRule        = "channelauthrule"
	ControllerAuthorityRecord        = "authorityrecord"
	ControllerQueueManagerConnection = "queuemanagerconnection"
)

// MQ operation names for mqweb adapter metrics.
const (
	MQOpPing              = "ping"
	MQOpGetQueue          = "get_queue"
	MQOpDefineQueue       = "define_queue"
	MQOpDeleteQueue       = "delete_queue"
	MQOpGetTopic          = "get_topic"
	MQOpDefineTopic       = "define_topic"
	MQOpDeleteTopic       = "delete_topic"
	MQOpGetChannel        = "get_channel"
	MQOpDefineChannel     = "define_channel"
	MQOpDeleteChannel     = "delete_channel"
	MQOpSetChannelAuth    = "set_channel_auth"
	MQOpGetChannelAuth    = "get_channel_auth"
	MQOpDeleteChannelAuth = "delete_channel_auth"
	MQOpSetAuthority      = "set_authority"
	MQOpGetAuthority      = "get_authority"
	MQOpDeleteAuthority   = "delete_authority"
	MQOpRunMQSC           = "run_mqsc"
)

var (
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mkurator_reconcile_total",
			Help: "Total reconciliations by controller and result.",
		},
		[]string{labelController, labelResult},
	)

	ReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mkurator_reconcile_errors_total",
			Help: "Total reconcile passes that returned an error to the manager.",
		},
		[]string{labelController},
	)

	MQOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mkurator_mq_operations_total",
			Help: "Total mqweb operations by operation and result.",
		},
		[]string{labelOperation, labelResult},
	)

	DriftDetectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mkurator_drift_detected_total",
			Help: "Total workload reconciles that detected attribute drift on IBM MQ.",
		},
		[]string{labelController},
	)

	CircuitBreakerTransitionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mkurator_mq_circuit_breaker_transitions_total",
			Help: "mqweb circuit breaker state transitions per connection client.",
		},
		[]string{"from", "to"},
	)
)

func init() {
	ctrmetrics.Registry.MustRegister(
		ReconcileTotal,
		ReconcileErrors,
		MQOperationsTotal,
		DriftDetectedTotal,
		CircuitBreakerTransitionsTotal,
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

// RecordDriftDetected increments drift detection counters for a workload controller.
func RecordDriftDetected(controller string) {
	DriftDetectedTotal.WithLabelValues(controller).Inc()
}

func RecordCircuitBreakerTransition(from, to string) {
	CircuitBreakerTransitionsTotal.WithLabelValues(from, to).Inc()
}
