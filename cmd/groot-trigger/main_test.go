package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
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

// TestListenConfigDisablesMPTCP guards the CVE-2026-98432 mitigation: Go 1.24+
// enables MPTCP on Linux listeners by default and the public exploit targets
// hosts exposing an MPTCP listener.
func TestListenConfigDisablesMPTCP(t *testing.T) {
	if listenConfig().MultipathTCP() {
		t.Fatal("listen config must not use MPTCP")
	}
}

// TestListenerDoesNotUseMPTCP verifies the production listener never negotiates
// MPTCP. The check is differential: a control listener with MPTCP explicitly
// enabled has to report MPTCP on the connection, otherwise the host kernel has
// no MPTCP support and the assertion below would be vacuous.
func TestListenerDoesNotUseMPTCP(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("MPTCP listeners exist only on Linux")
	}
	ctx := context.Background()

	var controlLC net.ListenConfig
	controlLC.SetMultipathTCP(true)
	control, err := controlLC.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("MPTCP control listener unavailable: %v", err)
	}
	defer func() { _ = control.Close() }()

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	dialer.SetMultipathTCP(true)

	controlConn, err := dialer.DialContext(ctx, "tcp", control.Addr().String())
	if err != nil {
		t.Skipf("MPTCP control dial failed: %v", err)
	}
	controlMPTCP := connUsesMPTCP(t, controlConn)
	_ = controlConn.Close()

	if !controlMPTCP {
		t.Skip("control connection did not negotiate MPTCP; differential check not applicable on this host")
	}

	l, err := listenConfig().Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()

	conn, err := dialer.DialContext(ctx, "tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	if connUsesMPTCP(t, conn) {
		t.Fatal("production listener negotiated MPTCP; want plain TCP")
	}
}

// connUsesMPTCP reports whether the connection is running over MPTCP.
func connUsesMPTCP(t *testing.T, c net.Conn) bool {
	t.Helper()
	tc, ok := c.(*net.TCPConn)
	if !ok {
		t.Fatalf("want *net.TCPConn, got %T", c)
	}
	use, err := tc.MultipathTCP()
	if err != nil {
		t.Fatalf("MultipathTCP: %v", err)
	}
	return use
}
