package mqpcf_test

import (
	"context"
	"strings"
	"testing"

	"github.com/konih/kurator/internal/adapter/mqpcf"
	"github.com/konih/kurator/internal/mqadmin"
)

func TestClient_ImplementsAdmin(t *testing.T) {
	t.Parallel()
	var _ mqadmin.Admin = (*mqpcf.Client)(nil)
}

func TestNewClient_RequiresQueueManager(t *testing.T) {
	t.Parallel()
	_, err := mqpcf.NewClient(mqpcf.Config{})
	if err == nil {
		t.Fatal("expected error when queue manager is empty")
	}
}

func TestClient_Ping_NotImplemented(t *testing.T) {
	t.Parallel()
	c, err := mqpcf.NewClient(mqpcf.Config{QueueManager: "QM1"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("unexpected error: %v", err)
	}
}
