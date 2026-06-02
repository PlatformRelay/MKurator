package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Setup configures controller-runtime to log through logr backed by slog.
func Setup(cfg Config) error {
	return SetupWithWriter(cfg, logWriter(cfg.Format))
}

// SetupWithWriter is like Setup but writes to w (used in tests).
func SetupWithWriter(cfg Config, w io.Writer) error {
	handler, err := NewHandler(cfg, w)
	if err != nil {
		return err
	}
	ctrl.SetLogger(logr.FromSlogHandler(handler))
	return nil
}

// NewHandler builds the root slog handler for the given config.
func NewHandler(cfg Config, w io.Writer) (slog.Handler, error) {
	level, err := cfg.slogLevel()
	if err != nil {
		return nil, err
	}
	var levelVar slog.LevelVar
	levelVar.Set(level)

	opts := &slog.HandlerOptions{
		Level: &levelVar,
	}
	var base slog.Handler
	switch cfg.Format {
	case FormatJSON:
		base = slog.NewJSONHandler(w, opts)
	case FormatText:
		base = slog.NewTextHandler(w, opts)
	default:
		return nil, fmt.Errorf("unsupported log format %q", cfg.Format)
	}
	return newRedactHandler(base), nil
}

// NewLogger returns a logr.Logger writing to w (used in tests).
func NewLogger(cfg Config, w io.Writer) (logr.Logger, error) {
	handler, err := NewHandler(cfg, w)
	if err != nil {
		return logr.Logger{}, err
	}
	return logr.FromSlogHandler(handler), nil
}

func logWriter(format Format) io.Writer {
	if format == FormatText {
		return os.Stderr
	}
	return os.Stdout
}

func (c Config) slogLevel() (slog.Level, error) {
	switch c.Level {
	case LevelDebug:
		return slog.LevelDebug, nil
	case LevelInfo:
		return slog.LevelInfo, nil
	case LevelWarn:
		return slog.LevelWarn, nil
	case LevelError:
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unsupported log level %q", c.Level)
	}
}
