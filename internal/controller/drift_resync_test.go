package controller

import (
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func TestSetDriftResyncInterval(t *testing.T) {
	t.Cleanup(func() {
		SetDriftResyncInterval(defaultDriftResyncLower, defaultDriftResyncUpper)
		SetDriftResyncRand(nil)
	})

	SetDriftResyncInterval(2*time.Minute, 8*time.Minute)
	if got := DriftResyncLower(); got != 2*time.Minute {
		t.Fatalf("lower = %v", got)
	}
	if got := DriftResyncUpper(); got != 8*time.Minute {
		t.Fatalf("upper = %v", got)
	}

	SetDriftResyncInterval(10*time.Minute, 5*time.Minute)
	if got := DriftResyncLower(); got != 2*time.Minute {
		t.Fatalf("invalid lower/upper should keep previous values, lower = %v", got)
	}
	if got := DriftResyncUpper(); got != 8*time.Minute {
		t.Fatalf("invalid lower/upper should keep previous values, upper = %v", got)
	}

	SetDriftResyncInterval(0, 5*time.Minute)
	if got := DriftResyncLower(); got != 2*time.Minute {
		t.Fatalf("zero lower should keep previous values, lower = %v", got)
	}
}

func TestDriftResyncAfter_JitterWithinBounds(t *testing.T) {
	t.Cleanup(func() {
		SetDriftResyncInterval(defaultDriftResyncLower, defaultDriftResyncUpper)
		SetDriftResyncRand(nil)
	})

	SetDriftResyncInterval(5*time.Minute, 10*time.Minute)
	SetDriftResyncRand(func() float64 { return 0 })
	if got := DriftResyncAfter(); got != 5*time.Minute {
		t.Fatalf("rand=0: got %v want 5m", got)
	}

	SetDriftResyncRand(func() float64 { return 1 })
	if got := DriftResyncAfter(); got != 10*time.Minute {
		t.Fatalf("rand=1: got %v want 10m", got)
	}

	SetDriftResyncRand(func() float64 { return 0.5 })
	if got := DriftResyncAfter(); got != 7*time.Minute+30*time.Second {
		t.Fatalf("rand=0.5: got %v want 7m30s", got)
	}
}

func TestWorkloadDriftResyncResult(t *testing.T) {
	t.Cleanup(func() {
		SetDriftResyncInterval(defaultDriftResyncLower, defaultDriftResyncUpper)
		SetDriftResyncRand(nil)
	})

	SetDriftResyncInterval(3*time.Minute, 3*time.Minute)
	result := workloadDriftResyncResult()
	if result != (ctrl.Result{RequeueAfter: 3 * time.Minute}) {
		t.Fatalf("result = %+v", result)
	}
}

func assertDriftResyncRequeue(t *testing.T, result ctrl.Result) {
	t.Helper()
	lower, upper := DriftResyncLower(), DriftResyncUpper()
	if result.RequeueAfter < lower || result.RequeueAfter > upper {
		t.Fatalf("RequeueAfter = %v, want [%v, %v]", result.RequeueAfter, lower, upper)
	}
}
