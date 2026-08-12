// Package main is the groot-trigger entrypoint.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hrodrig/groot-trigger/internal/config"
	"github.com/hrodrig/groot-trigger/internal/jobs"
	"github.com/hrodrig/groot-trigger/internal/logging"
	"github.com/hrodrig/groot-trigger/internal/proxy"
	"github.com/hrodrig/groot-trigger/internal/ratelimit"
	"github.com/hrodrig/groot-trigger/internal/server"
)

// Set via -ldflags at build time (Makefile / GoReleaser).
var (
	version   = "dev"
	commit    = "unknown"
	branch    = "unknown"
	buildDate = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "version", "--version", "-V":
			fmt.Printf("groot-trigger %s commit=%s branch=%s date=%s\n", version, commit, branch, buildDate)
			return 0
		}
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "groot-trigger: %v\n", err)
		return 1
	}
	logging.Setup(cfg.LogFormat, cfg.LogLevel)

	fmt.Fprintf(os.Stderr, "groot-trigger %s | build %s | listen %s | api_key %s\n",
		version, buildDate, cfg.ListenAddr, config.MaskSecret(cfg.APIKey))

	httpSrv, err := newHTTPServer(cfg, jobs.NewInCluster)
	if err != nil {
		slog.Error("setup failed", "error", err)
		return 1
	}
	return listenAndServe(httpSrv)
}

// newInClusterFn is overridable in tests.
type newInClusterFn func(config.Config) (*jobs.K8sStarter, error)

func newHTTPServer(cfg config.Config, newKS newInClusterFn) (*http.Server, error) {
	trusted := proxy.ParseTrustedProxies(cfg.TrustedProxies)
	if trusted.Empty() && cfg.RateLimitPost.Requests > 0 {
		slog.Warn("Client IP headers ignored: GROOT_TRIGGER_TRUSTED_PROXIES empty (safe for ClusterIP)")
	}

	var starter jobs.Starter
	readyOK := true
	ks, err := newKS(cfg)
	if err != nil {
		slog.Warn("kubernetes client unavailable; /readyz fails until in-cluster config works", "error", err)
		readyOK = false
		starter = jobs.Unavailable(err)
	} else {
		starter = ks
	}

	srv := &server.Server{
		Cfg:     cfg,
		Jobs:    starter,
		Limit:   ratelimit.New(cfg.RateLimitPost, cfg.RateLimitGlobal, trusted),
		Trusted: trusted,
		Ready:   func() bool { return readyOK },
	}

	slog.Info("starting",
		"image", cfg.GrootImage,
		"namespace", cfg.GrootNamespace,
		"rate_limit_post", fmt.Sprintf("%d/%s", cfg.RateLimitPost.Requests, cfg.RateLimitPost.Window),
		"trusted_proxies", !trusted.Empty(),
	)

	return &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}

func listenAndServe(httpSrv *http.Server) int {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("listen", "addr", httpSrv.Addr)
		errCh <- httpSrv.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
		return 0
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server stopped", "error", err)
			return 1
		}
		return 0
	}
}
