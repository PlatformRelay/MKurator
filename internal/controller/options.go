package controller

import "sigs.k8s.io/controller-runtime/pkg/controller"

var maxConcurrentReconciles = 1

// SetMaxConcurrentReconciles configures worker count for all Kurator controllers (minimum 1).
func SetMaxConcurrentReconciles(n int) {
	if n < 1 {
		n = 1
	}
	maxConcurrentReconciles = n
}

func controllerOptions() controller.Options {
	return controller.Options{MaxConcurrentReconciles: maxConcurrentReconciles}
}
