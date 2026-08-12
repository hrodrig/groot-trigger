// Package config loads groot-trigger settings from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime settings for the trigger process and Job template.
type Config struct {
	ListenAddr string
	APIKey     string

	LogFormat string // json | text
	LogLevel  string // debug|info|warn|error

	TrustedProxies string

	RateLimitPost   LimitSpec
	RateLimitGlobal LimitSpec

	GrootImage         string
	GrootNamespace     string
	GrootConfigMap     string
	GrootConfigKey     string
	GrootOutPVC        string
	GrootJobSA         string
	GrootExtraArgs     []string
	GrootEnvFromSecret string

	JobTTLSeconds int32
}

// LimitSpec is requests per window (0 = disabled).
type LimitSpec struct {
	Requests int
	Window   time.Duration
}

// LoadFromEnv reads configuration. Returns error if API key is empty (fail closed).
func LoadFromEnv() (Config, error) {
	cfg := Config{
		ListenAddr:         envOr("LISTEN_ADDR", ":8080"),
		APIKey:             strings.TrimSpace(os.Getenv("GROOT_TRIGGER_API_KEY")),
		LogFormat:          strings.ToLower(envOr("GROOT_TRIGGER_LOG_FORMAT", "json")),
		LogLevel:           strings.ToLower(envOr("GROOT_TRIGGER_LOG_LEVEL", "info")),
		TrustedProxies:     strings.TrimSpace(os.Getenv("GROOT_TRIGGER_TRUSTED_PROXIES")),
		GrootImage:         envOr("GROOT_IMAGE", "ghcr.io/hrodrig/groot:v1.1.1"),
		GrootNamespace:     strings.TrimSpace(os.Getenv("GROOT_NAMESPACE")),
		GrootConfigMap:     envOr("GROOT_CONFIGMAP", "groot-config"),
		GrootConfigKey:     envOr("GROOT_CONFIG_KEY", "groot.yml"),
		GrootOutPVC:        strings.TrimSpace(os.Getenv("GROOT_OUT_PVC")),
		GrootJobSA:         envOr("GROOT_JOB_SA", "groot"),
		GrootEnvFromSecret: strings.TrimSpace(os.Getenv("GROOT_ENVFROM_SECRET")),
		JobTTLSeconds:      3600,
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("GROOT_TRIGGER_API_KEY is required (fail closed)")
	}
	if extra := strings.TrimSpace(os.Getenv("GROOT_EXTRA_ARGS")); extra != "" {
		cfg.GrootExtraArgs = strings.Fields(extra)
	}
	var err error
	cfg.RateLimitPost, err = ParseLimitSpec(envOr("GROOT_TRIGGER_RATE_LIMIT_POST", "10/1m"))
	if err != nil {
		return Config{}, fmt.Errorf("GROOT_TRIGGER_RATE_LIMIT_POST: %w", err)
	}
	cfg.RateLimitGlobal, err = ParseLimitSpec(envOr("GROOT_TRIGGER_RATE_LIMIT_GLOBAL", "0"))
	if err != nil {
		return Config{}, fmt.Errorf("GROOT_TRIGGER_RATE_LIMIT_GLOBAL: %w", err)
	}
	return cfg, nil
}

// ParseLimitSpec parses "N/1m", "N/1h", "N/30s", or "0" / empty (disabled).
func ParseLimitSpec(s string) (LimitSpec, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || s == "0" || s == "off" || s == "disabled" {
		return LimitSpec{}, nil
	}
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return LimitSpec{}, fmt.Errorf("invalid limit %q (want N/duration e.g. 10/1m)", s)
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil || n < 0 {
		return LimitSpec{}, fmt.Errorf("invalid request count in %q", s)
	}
	if n == 0 {
		return LimitSpec{}, nil
	}
	d, err := time.ParseDuration(parts[1])
	if err != nil || d <= 0 {
		return LimitSpec{}, fmt.Errorf("invalid duration in %q", s)
	}
	return LimitSpec{Requests: n, Window: d}, nil
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// MaskSecret returns a partially redacted secret for startup banners.
func MaskSecret(s string) string {
	r := []rune(s)
	if len(r) < 8 {
		return "[masked]"
	}
	return string(r[:4]) + "...." + string(r[len(r)-4:])
}
