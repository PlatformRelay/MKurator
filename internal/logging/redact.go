package logging

import (
	"context"
	"log/slog"
	"strings"
)

var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"token":         {},
	"authorization": {},
	"secret":        {},
	"credentials":   {},
	"csrf":          {},
	"apikey":        {},
	"api_key":       {},
}

// redactHandler masks sensitive attribute keys before writing log records.
type redactHandler struct {
	next slog.Handler
}

func newRedactHandler(next slog.Handler) *redactHandler {
	return &redactHandler{next: next}
}

func (h *redactHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *redactHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.NumAttrs() == 0 {
		return h.next.Handle(ctx, record)
	}
	attrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, redactAttr(attr))
		return true
	})
	cloned := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	cloned.AddAttrs(attrs...)
	return h.next.Handle(ctx, cloned)
}

func (h *redactHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		out[i] = redactAttr(attr)
	}
	return &redactHandler{next: h.next.WithAttrs(out)}
}

func (h *redactHandler) WithGroup(name string) slog.Handler {
	return &redactHandler{next: h.next.WithGroup(name)}
}

func redactAttr(attr slog.Attr) slog.Attr {
	key := strings.ToLower(attr.Key)
	if _, ok := sensitiveKeys[key]; ok {
		return slog.String(attr.Key, "[REDACTED]")
	}
	return attr
}
