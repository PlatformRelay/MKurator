package logging

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	EnvConfig = "KURATOR_LOG_CONFIG"
	EnvLevel  = "KURATOR_LOG_LEVEL"
	EnvFormat = "KURATOR_LOG_FORMAT"
)

// Level is the minimum log level emitted to the sink.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Format selects JSON (production) or text (local) output.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Config holds resolved logging settings.
type Config struct {
	Level  Level  `json:"level"  yaml:"level"`
	Format Format `json:"format" yaml:"format"`
}

// Options supplies overrides from CLI flags. Empty strings mean "not set".
type Options struct {
	ConfigPath string
	Level      string
	Format     string
}

// Load merges defaults, config file, environment variables, and CLI flags.
// Later sources win: defaults < file < env < flags.
func Load(opts Options) (Config, error) {
	cfg := defaultConfig()

	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = os.Getenv(EnvConfig)
	}
	if configPath != "" {
		fileCfg, err := loadFile(configPath)
		if err != nil {
			return Config{}, fmt.Errorf("load logging config %q: %w", configPath, err)
		}
		cfg = merge(cfg, fileCfg)
	}

	if v := os.Getenv(EnvLevel); v != "" {
		level, err := parseLevel(v)
		if err != nil {
			return Config{}, err
		}
		cfg.Level = level
	}
	if v := os.Getenv(EnvFormat); v != "" {
		format, err := parseFormat(v)
		if err != nil {
			return Config{}, err
		}
		cfg.Format = format
	}

	if opts.Level != "" {
		level, err := parseLevel(opts.Level)
		if err != nil {
			return Config{}, err
		}
		cfg.Level = level
	}
	if opts.Format != "" {
		format, err := parseFormat(opts.Format)
		if err != nil {
			return Config{}, err
		}
		cfg.Format = format
	}

	return cfg, nil
}

func defaultConfig() Config {
	format := FormatText
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		format = FormatJSON
	}
	return Config{
		Level:  LevelInfo,
		Format: format,
	}
}

func loadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Level != "" {
		level, err := parseLevel(string(cfg.Level))
		if err != nil {
			return Config{}, err
		}
		cfg.Level = level
	}
	if cfg.Format != "" {
		format, err := parseFormat(string(cfg.Format))
		if err != nil {
			return Config{}, err
		}
		cfg.Format = format
	}
	return cfg, nil
}

func merge(base, overlay Config) Config {
	if overlay.Level != "" {
		base.Level = overlay.Level
	}
	if overlay.Format != "" {
		base.Format = overlay.Format
	}
	return base
}

func parseLevel(raw string) (Level, error) {
	switch Level(strings.ToLower(strings.TrimSpace(raw))) {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return Level(strings.ToLower(strings.TrimSpace(raw))), nil
	default:
		return "", fmt.Errorf("invalid log level %q (want debug, info, warn, or error)", raw)
	}
}

func parseFormat(raw string) (Format, error) {
	switch Format(strings.ToLower(strings.TrimSpace(raw))) {
	case FormatJSON, FormatText:
		return Format(strings.ToLower(strings.TrimSpace(raw))), nil
	default:
		return "", fmt.Errorf("invalid log format %q (want json or text)", raw)
	}
}
