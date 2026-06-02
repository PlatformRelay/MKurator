package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/konradheimel/kurator/internal/logging"
)

func TestRedactHandlerWithSlogLogger(t *testing.T) {
	var buf bytes.Buffer
	handler, err := logging.NewHandler(logging.Config{
		Level:  logging.LevelInfo,
		Format: logging.FormatJSON,
	}, &buf)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	slog.New(handler).Info("event", slog.String("token", "abc123"))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry["token"] != "[REDACTED]" {
		t.Fatalf("token: got %v", entry["token"])
	}
}
