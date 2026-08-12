// Package main is the groot-trigger entrypoint.
package main

import (
	"context"
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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-V":
			fmt.Printf("groot-trigger %s commit=%s branch=%s date=%s\n", version, commit, branch, buildDate)
			return
		}
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "groot-trigger: %v\n", err)
		os.Exit(1)
	}
	logging.Setup(cfg.LogFormat, cfg.LogLevel)

	fmt.Fprintf(os.Stderr, "groot-trigger %s | build %s | listen %s | api_key %s\n",
		version, buildDate, cfg.ListenAddr, maskSecret(cfg.APIKey))

	trusted := proxy.ParseTrustedProxies(cfg.TrustedProxies)
	if trusted.Empty() && cfg.RateLimitPost.Requests > 0 {
		slog.Warn("Client IP headers ignored: GROOT_TRIGGER_TRUSTED_PROXIES empty (safe for ClusterIP)")
	}

	var starter jobs.Starter
	readyOK := true
	ks, err := jobs.NewInCluster(cfg)
	if err != nil {
		slog.Warn("kubernetes client unavailable; /readyz fails until in-cluster config works", "error", err)
		readyOK = false
		starter = unavailableStarter{err: err}
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

	httpSrv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("listen", "addr", cfg.ListenAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

func maskSecret(s string) string {
	r := []rune(s)
	if len(r) < 8 {
		return "[masked]"
	}
	return string(r[:4]) + "...." + string(r[len(r)-4:])
}

type unavailableStarter struct{ err error }

func (u unavailableStarter) ActiveJob(context.Context) (string, bool, error) {
	return "", false, u.err
}

func (u unavailableStarter) Create(context.Context, string, string) (jobs.Result, error) {
	return jobs.Result{}, u.err
}
