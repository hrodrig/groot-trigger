// Package server implements the HTTP API and vanilla collect UI.
package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hrodrig/groot-trigger/internal/auth"
	"github.com/hrodrig/groot-trigger/internal/config"
	"github.com/hrodrig/groot-trigger/internal/jobs"
	"github.com/hrodrig/groot-trigger/internal/proxy"
	"github.com/hrodrig/groot-trigger/internal/ratelimit"
)

// Server is the HTTP front-end.
type Server struct {
	Cfg     config.Config
	Jobs    jobs.Starter
	Limit   *ratelimit.Limiter
	Trusted *proxy.TrustedProxies
	Ready   func() bool
}

// Handler returns the root mux with middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	mux.HandleFunc("GET /v1/collect", s.handleCollectGET)
	mux.HandleFunc("POST /v1/collect", s.handleCollectPOST)
	return s.accessLog(mux)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok\n")
}

func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	if s.Ready != nil && !s.Ready() {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok\n")
}

func (s *Server) handleCollectGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = formTmpl.Execute(w, nil)
}

func (s *Server) handleCollectPOST(w http.ResponseWriter, r *http.Request) {
	wantJSON := wantsJSON(r)

	if s.Limit != nil && !s.Limit.Allow(r) {
		s.respond(w, r, wantJSON, http.StatusTooManyRequests, map[string]any{"error": "rate_limited"}, "Too many requests")
		return
	}
	key := auth.ExtractKey(r)
	if !auth.Valid(key, s.Cfg.APIKey) {
		s.respond(w, r, wantJSON, http.StatusUnauthorized, map[string]any{"error": "unauthorized"}, "Unauthorized")
		return
	}

	var message string
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		var body struct {
			Message string `json:"message"`
		}
		dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
		if err := dec.Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			s.respond(w, r, wantJSON, http.StatusBadRequest, map[string]any{"error": "bad_request"}, "Bad request")
			return
		}
		message = body.Message
	} else {
		_ = r.ParseForm()
		message = r.Form.Get("message")
	}

	runID, err := newRunID()
	if err != nil {
		s.respond(w, r, wantJSON, http.StatusInternalServerError, map[string]any{"error": "internal", "detail": "run_id"}, "Internal error")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	res, err := s.Jobs.Create(ctx, runID, message)
	if err != nil {
		var busy *jobs.ErrBusy
		if errors.As(err, &busy) {
			s.respond(w, r, wantJSON, http.StatusConflict, map[string]any{"error": "collect_in_progress", "job": busy.JobName}, "Collect already in progress")
			return
		}
		slog.Error("job create failed", "error", err, "run_id", runID)
		s.respond(w, r, wantJSON, http.StatusInternalServerError, map[string]any{"error": "internal", "detail": "job_create"}, "Internal error")
		return
	}
	s.respond(w, r, wantJSON, http.StatusAccepted, map[string]any{"run_id": res.RunID, "job": res.JobName}, "Collect started")
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, asJSON bool, code int, payload map[string]any, htmlTitle string) {
	if asJSON {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(payload)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	data := map[string]any{
		"Title": htmlTitle,
		"Code":  code,
		"Body":  payload,
	}
	_ = resultTmpl.Execute(w, data)
}

func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return true
	}
	if r.Header.Get("X-API-Key") != "" || strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		return true
	}
	return false
}

func newRunID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		level := slog.LevelInfo
		switch {
		case rec.status >= 500:
			level = slog.LevelError
		case rec.status >= 400:
			level = slog.LevelWarn
		}
		slog.Log(r.Context(), level, "http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"ip", proxy.ClientIP(r, s.Trusted),
			"dur", time.Since(start).Round(time.Millisecond),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

var formTmpl = template.Must(template.New("form").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>GROOT trigger</title>
<style>
:root { --bg:#f4f6f8; --fg:#1a1f24; --accent:#0b7a4b; --muted:#5c6b73; --border:#d0d7de; --mono: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; --sans: "IBM Plex Sans", "Segoe UI", sans-serif; }
* { box-sizing: border-box; }
html, body { margin:0; min-height:100%; background: radial-gradient(1200px 600px at 10% -10%, #e8f5ee, transparent), var(--bg); color:var(--fg); font-family:var(--sans); }
main { max-width: 28rem; margin: 12vh auto 2rem; padding: 0 1.25rem; }
h1 { font-size: 1.75rem; font-weight: 650; letter-spacing: -0.02em; margin: 0 0 .35rem; }
p { color: var(--muted); margin: 0 0 1.5rem; line-height: 1.45; }
label { display:block; font-size:.85rem; margin-bottom:.35rem; color:var(--muted); }
input[type=password], input[type=text] { width:100%; padding:.7rem .8rem; border:1px solid var(--border); border-radius:6px; font: inherit; background:#fff; margin-bottom:1rem; }
button { width:100%; padding:.85rem 1rem; border:0; border-radius:6px; background:var(--accent); color:#fff; font-weight:600; font-size:1rem; cursor:pointer; }
button:hover { filter: brightness(1.05); }
.foot { margin-top:1.25rem; font-family:var(--mono); font-size:.75rem; color:var(--muted); }
</style>
</head>
<body>
<main>
  <h1>GROOT trigger</h1>
  <p>Start an in-cluster collect Job. Completion appears via notify or object storage.</p>
  <form method="POST" action="/v1/collect">
    <label for="api_key">API key</label>
    <input id="api_key" name="api_key" type="password" autocomplete="current-password" required>
    <label for="message">Message (optional)</label>
    <input id="message" name="message" type="text" maxlength="200">
    <button type="submit">Generate GROOT files</button>
  </form>
  <p class="foot">POST /v1/collect · fire-and-forget</p>
</main>
</body>
</html>`))

var resultTmpl = template.Must(template.New("result").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
body { font-family: "IBM Plex Sans", "Segoe UI", sans-serif; background:#f4f6f8; color:#1a1f24; margin:0; }
main { max-width:28rem; margin:12vh auto; padding:0 1.25rem; }
h1 { font-size:1.4rem; }
pre { background:#fff; border:1px solid #d0d7de; padding:1rem; border-radius:6px; overflow:auto; font-size:.85rem; }
a { color:#0b7a4b; }
</style>
</head>
<body>
<main>
  <h1>{{.Title}} ({{.Code}})</h1>
  <pre>{{printf "%+v" .Body}}</pre>
  <p><a href="/v1/collect">Back</a></p>
</main>
</body>
</html>`))
