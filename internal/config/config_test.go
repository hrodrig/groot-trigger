package config

import (
	"testing"
	"time"
)

func TestParseLimitSpec(t *testing.T) {
	cases := []struct {
		in      string
		wantN   int
		wantDur time.Duration
		wantErr bool
	}{
		{"10/1m", 10, time.Minute, false},
		{"0", 0, 0, false},
		{"off", 0, 0, false},
		{"", 0, 0, false},
		{"bad", 0, 0, true},
		{"10", 0, 0, true},
	}
	for _, tc := range cases {
		got, err := ParseLimitSpec(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("%q: want error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got.Requests != tc.wantN || got.Window != tc.wantDur {
			t.Fatalf("%q: got %+v", tc.in, got)
		}
	}
}

func TestLoadFromEnvFailClosed(t *testing.T) {
	t.Setenv("GROOT_TRIGGER_API_KEY", "")
	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("expected error when API key empty")
	}
}

func TestLoadFromEnvOK(t *testing.T) {
	t.Setenv("GROOT_TRIGGER_API_KEY", "secret")
	t.Setenv("LISTEN_ADDR", ":9090")
	t.Setenv("GROOT_TRIGGER_RATE_LIMIT_POST", "5/1m")
	t.Setenv("GROOT_EXTRA_ARGS", "--verbose")
	t.Setenv("GROOT_TRIGGER_RATE_LIMIT_GLOBAL", "30/1m")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "secret" || cfg.ListenAddr != ":9090" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if cfg.RateLimitPost.Requests != 5 {
		t.Fatalf("rate: %+v", cfg.RateLimitPost)
	}
	if cfg.RateLimitGlobal.Requests != 30 {
		t.Fatalf("global: %+v", cfg.RateLimitGlobal)
	}
	if len(cfg.GrootExtraArgs) != 1 || cfg.GrootExtraArgs[0] != "--verbose" {
		t.Fatalf("extra: %#v", cfg.GrootExtraArgs)
	}
}

func TestLoadFromEnvBadRate(t *testing.T) {
	t.Setenv("GROOT_TRIGGER_API_KEY", "secret")
	t.Setenv("GROOT_TRIGGER_RATE_LIMIT_POST", "nope")
	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("expected error")
	}
}

func TestMaskSecret(t *testing.T) {
	if MaskSecret("short") != "[masked]" {
		t.Fatal("short")
	}
	if got := MaskSecret("abcdefghij"); got != "abcd....ghij" {
		t.Fatal(got)
	}
}
