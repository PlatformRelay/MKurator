package controller

import (
	"testing"

	"github.com/konradheimel/kurator/internal/mqadmin"
)

func TestNeedsUpdate(t *testing.T) {
	t.Parallel()
	desired := mqadmin.QueueSpec{
		Name: "APP.ORDERS",
		Attributes: map[string]string{
			"maxdepth": "5000",
		},
	}
	observed := &mqadmin.QueueState{
		Attributes: map[string]string{"maxdepth": "5000"},
	}
	if needsUpdate(desired, observed) {
		t.Fatal("expected no update when attributes match")
	}
	observed.Attributes["maxdepth"] = "1000"
	if !needsUpdate(desired, observed) {
		t.Fatal("expected update when maxdepth drifts")
	}
}
