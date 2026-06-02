package mqrest

import (
	"testing"

	"github.com/konradheimel/kurator/internal/mqadmin"
)

func TestDefineQueueParameters(t *testing.T) {
	t.Parallel()
	params := defineQueueParameters(mqadmin.QueueSpec{
		Name: "APP.ORDERS",
		Attributes: map[string]string{
			attrMaxDepth: "5000",
			"descr":      "orders",
		},
	})
	if params["replace"] != "yes" {
		t.Fatalf("replace = %v", params["replace"])
	}
	if params[attrMaxDepth] != 5000 {
		t.Fatalf("maxdepth should be int 5000, got %T(%v)", params[attrMaxDepth], params[attrMaxDepth])
	}
	if params["descr"] != "orders" {
		t.Fatalf("descr = %v", params["descr"])
	}
}

func TestQueueDisplayParametersExcludeMaxmsglen(t *testing.T) {
	t.Parallel()
	for _, p := range queueDisplayParameters {
		if p == "maxmsglen" {
			t.Fatal("maxmsglen must not be in display parameters for mqweb 9.4")
		}
	}
}
