package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hrodrig/groot-trigger/internal/config"
	"github.com/hrodrig/groot-trigger/internal/jobs"
	"github.com/hrodrig/groot-trigger/internal/proxy"
	"github.com/hrodrig/groot-trigger/internal/ratelimit"
	"time"
)

type fakeJobs struct {
	busyName string
	created  int
}

func (f *fakeJobs) ActiveJob(context.Context) (string, bool, error) {
	if f.busyName != "" {
		return f.busyName, true, nil
	}
	return "", false, nil
}

func (f *fakeJobs) Create(_ context.Context, runID, _ string) (jobs.Result, error) {
	if f.busyName != "" {
		return jobs.Result{}, &jobs.ErrBusy{JobName: f.busyName}
	}
	f.created++
	f.busyName = "groot-collect-" + runID[:8]
	return jobs.Result{RunID: runID, JobName: f.busyName}, nil
}

func testServer(fj jobs.Starter, postLimit config.LimitSpec) *Server {
	cfg := config.Config{APIKey: "secret"}
	tp := proxy.ParseTrustedProxies("")
	return &Server{
		Cfg:     cfg,
		Jobs:    fj,
		Limit:   ratelimit.New(postLimit, config.LimitSpec{}, tp),
		Trusted: tp,
		Ready:   func() bool { return true },
	}
}

func TestHealthz(t *testing.T) {
	s := testServer(&fakeJobs{}, config.LimitSpec{})
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != 200 {
		t.Fatal(rr.Code)
	}
}

func TestCollectGETHasButton(t *testing.T) {
	s := testServer(&fakeJobs{}, config.LimitSpec{})
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/collect", nil))
	body := rr.Body.String()
	if !strings.Contains(body, "Generate GROOT files") || !strings.Contains(body, `name="api_key"`) {
		t.Fatal(body)
	}
}

func TestCollectUnauthorized(t *testing.T) {
	s := testServer(&fakeJobs{}, config.LimitSpec{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", nil)
	req.Header.Set("Accept", "application/json")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatal(rr.Code)
	}
}

func TestCollectAccepted(t *testing.T) {
	fj := &fakeJobs{}
	s := testServer(fj, config.LimitSpec{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", nil)
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("Accept", "application/json")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("code %d body %s", rr.Code, rr.Body.String())
	}
	var got map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &got)
	if got["run_id"] == nil || got["job"] == nil {
		t.Fatal(got)
	}
}

func TestCollectBusy(t *testing.T) {
	fj := &fakeJobs{busyName: "existing"}
	s := testServer(fj, config.LimitSpec{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", nil)
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("Accept", "application/json")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatal(rr.Code)
	}
}

func TestCollectRateLimited(t *testing.T) {
	s := testServer(&fakeJobs{}, config.LimitSpec{Requests: 1, Window: time.Minute})
	h := s.Handler()
	mk := func() *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/collect", nil)
		req.RemoteAddr = "198.51.100.1:9"
		req.Header.Set("X-API-Key", "secret")
		req.Header.Set("Accept", "application/json")
		h.ServeHTTP(rr, req)
		return rr
	}
	if mk().Code != http.StatusAccepted {
		t.Fatal("first")
	}
	if mk().Code != http.StatusTooManyRequests {
		t.Fatal("second should 429")
	}
}
