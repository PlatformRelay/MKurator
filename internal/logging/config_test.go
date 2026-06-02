package logging_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/konradheimel/kurator/internal/logging"
)

func TestLoadDefaultsLocal(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv(logging.EnvConfig, "")
	t.Setenv(logging.EnvLevel, "")
	t.Setenv(logging.EnvFormat, "")

	cfg, err := logging.Load(logging.Options{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Level != logging.LevelInfo {
		t.Fatalf("level: got %q want info", cfg.Level)
	}
	if cfg.Format != logging.FormatText {
		t.Fatalf("format: got %q want text", cfg.Format)
	}
}

func TestLoadDefaultsInCluster(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
	t.Setenv(logging.EnvConfig, "")
	t.Setenv(logging.EnvLevel, "")
	t.Setenv(logging.EnvFormat, "")

	cfg, err := logging.Load(logging.Options{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Format != logging.FormatJSON {
		t.Fatalf("format: got %q want json", cfg.Format)
	}
}

func TestLoadPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "logging.yaml")
	if err := os.WriteFile(path, []byte("level: warn\nformat: text\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv(logging.EnvConfig, path)
	t.Setenv(logging.EnvLevel, "error")
	t.Setenv(logging.EnvFormat, "json")

	cfg, err := logging.Load(logging.Options{Level: "debug"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Level != logging.LevelDebug {
		t.Fatalf("level: got %q want debug (flag wins)", cfg.Level)
	}
	if cfg.Format != logging.FormatJSON {
		t.Fatalf("format: got %q want json (env wins over file)", cfg.Format)
	}
}

func TestLoadInvalidLevel(t *testing.T) {
	_, err := logging.Load(logging.Options{Level: "trace"})
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}
