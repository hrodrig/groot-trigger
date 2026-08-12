package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/hrodrig/groot-trigger/internal/config"
	"github.com/hrodrig/groot-trigger/internal/jobs"
)

func TestRunVersion(t *testing.T) {
	if code := run([]string{"version"}); code != 0 {
		t.Fatalf("code %d", code)
	}
	if code := run([]string{"--version"}); code != 0 {
		t.Fatalf("--version %d", code)
	}
	if code := run([]string{"-V"}); code != 0 {
		t.Fatalf("-V %d", code)
	}
}

func TestRunMissingAPIKey(t *testing.T) {
	_ = os.Unsetenv("GROOT_TRIGGER_API_KEY")
	t.Setenv("GROOT_TRIGGER_API_KEY", "")
	if code := run(nil); code != 1 {
		t.Fatalf("want 1 got %d", code)
	}
}

func TestNewHTTPServerUnavailableK8s(t *testing.T) {
	cfg := config.Config{
		ListenAddr:    "127.0.0.1:0",
		APIKey:        "secret",
		LogFormat:     "json",
		LogLevel:      "info",
		RateLimitPost: config.LimitSpec{Requests: 10, Window: time.Minute},
		GrootImage:    "ghcr.io/hrodrig/groot:v1.1.1",
	}
	srv, err := newHTTPServer(cfg, func(config.Config) (*jobs.K8sStarter, error) {
		return nil, errors.New("no cluster")
	})
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != 200 {
		t.Fatal(rr.Code)
	}
}

func TestNewHTTPServerWithStarter(t *testing.T) {
	cfg := config.Config{
		ListenAddr: "127.0.0.1:0",
		APIKey:     "secret",
		GrootImage: "img",
	}
	ks := &jobs.K8sStarter{NS: "ns", Cfg: cfg}
	srv, err := newHTTPServer(cfg, func(config.Config) (*jobs.K8sStarter, error) {
		return ks, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != 200 {
		t.Fatal(rr.Code)
	}
}

func TestListenAndServeShutdown(t *testing.T) {
	srv := &http.Server{Addr: "127.0.0.1:0", Handler: http.NewServeMux()}
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = srv.Shutdown(context.Background())
	}()
	if code := listenAndServe(srv); code != 0 {
		t.Fatalf("code %d", code)
	}
}
