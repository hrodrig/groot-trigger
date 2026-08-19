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
	busyName    string
	created     int
	lastMessage string
}

func (f *fakeJobs) ActiveJob(context.Context) (string, bool, error) {
	if f.busyName != "" {
		return f.busyName, true, nil
	}
	return "", false, nil
}

func (f *fakeJobs) Create(_ context.Context, runID, message string) (jobs.Result, error) {
	if f.busyName != "" {
		return jobs.Result{}, &jobs.ErrBusy{JobName: f.busyName}
	}
	f.created++
	f.lastMessage = message
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
		Version: "0.1.2",
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
	if !strings.Contains(body, `maxlength="128"`) {
		t.Fatal("expected message maxlength 128")
	}
	if !strings.Contains(body, "fire-and-forget · v0.1.2") {
		t.Fatal("expected version in footer")
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

func TestReadyz(t *testing.T) {
	s := testServer(&fakeJobs{}, config.LimitSpec{})
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != 200 {
		t.Fatal(rr.Code)
	}
	s.Ready = func() bool { return false }
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatal(rr.Code)
	}
}

func TestCollectHTMLFormAndResult(t *testing.T) {
	fj := &fakeJobs{}
	s := testServer(fj, config.LimitSpec{})
	rr := httptest.NewRecorder()
	body := strings.NewReader("api_key=secret&message=hi")
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("code %d body %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Collect started") {
		t.Fatal(rr.Body.String())
	}
	if fj.lastMessage != "hi" {
		t.Fatalf("message=%q", fj.lastMessage)
	}
}

func TestCollectJSONBody(t *testing.T) {
	fj := &fakeJobs{}
	s := testServer(fj, config.LimitSpec{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", strings.NewReader(`{"message":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("code %d %s", rr.Code, rr.Body.String())
	}
	if fj.lastMessage != "x" {
		t.Fatalf("message=%q", fj.lastMessage)
	}
}

func TestCollectMessageTooLong(t *testing.T) {
	s := testServer(&fakeJobs{}, config.LimitSpec{})
	long := strings.Repeat("a", maxMessageRunes+1)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", strings.NewReader(`{"message":"`+long+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("code %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "message_too_long") {
		t.Fatal(rr.Body.String())
	}
}

func TestCollectMessageMaxOK(t *testing.T) {
	fj := &fakeJobs{}
	s := testServer(fj, config.LimitSpec{})
	ok := strings.Repeat("b", maxMessageRunes)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", strings.NewReader(`{"message":"`+ok+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("code %d %s", rr.Code, rr.Body.String())
	}
	if fj.lastMessage != ok {
		t.Fatalf("len=%d", len(fj.lastMessage))
	}
}

func TestCollectBadJSON(t *testing.T) {
	s := testServer(&fakeJobs{}, config.LimitSpec{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", strings.NewReader(`{`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatal(rr.Code)
	}
}

type errJobs struct{ err error }

func (e errJobs) ActiveJob(context.Context) (string, bool, error) { return "", false, nil }
func (e errJobs) Create(context.Context, string, string) (jobs.Result, error) {
	return jobs.Result{}, e.err
}

func TestCollectCreateError(t *testing.T) {
	s := testServer(errJobs{err: context.DeadlineExceeded}, config.LimitSpec{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/collect", nil)
	req.Header.Set("X-API-Key", "secret")
	req.Header.Set("Accept", "application/json")
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatal(rr.Code)
	}
}

func TestWantsJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	if wantsJSON(r) {
		t.Fatal("plain")
	}
	r.Header.Set("Accept", "application/json")
	if !wantsJSON(r) {
		t.Fatal("accept")
	}
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Authorization", "Bearer x")
	if !wantsJSON(r) {
		t.Fatal("bearer")
	}
}

func TestVersionLabel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "dev"},
		{"dev", "dev"},
		{"0.1.2", "v0.1.2"},
		{"v0.1.2", "v0.1.2"},
	}
	for _, tc := range cases {
		s := &Server{Version: tc.in}
		if got := s.versionLabel(); got != tc.want {
			t.Fatalf("in=%q got=%q want=%q", tc.in, got, tc.want)
		}
	}
}
