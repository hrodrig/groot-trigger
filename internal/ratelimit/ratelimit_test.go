package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hrodrig/groot-trigger/internal/config"
	"github.com/hrodrig/groot-trigger/internal/proxy"
)

func TestPerIPLimit(t *testing.T) {
	l := New(config.LimitSpec{Requests: 2, Window: time.Minute}, config.LimitSpec{}, proxy.ParseTrustedProxies(""))
	r := httptest.NewRequest(http.MethodPost, "/v1/collect", nil)
	r.RemoteAddr = "192.0.2.1:1"
	if !l.Allow(r) {
		t.Fatal("first should pass")
	}
	if !l.Allow(r) {
		t.Fatal("second should pass")
	}
	if l.Allow(r) {
		t.Fatal("third should fail")
	}
}

func TestGlobalLimit(t *testing.T) {
	l := New(config.LimitSpec{}, config.LimitSpec{Requests: 1, Window: time.Minute}, proxy.ParseTrustedProxies(""))
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.RemoteAddr = "192.0.2.1:1"
	if !l.Allow(r) {
		t.Fatal("first")
	}
	if l.Allow(r) {
		t.Fatal("global should block")
	}
}

func TestNilLimiterAllows(t *testing.T) {
	var l *Limiter
	if !l.Allow(httptest.NewRequest(http.MethodGet, "/", nil)) {
		t.Fatal("nil")
	}
	l = New(config.LimitSpec{}, config.LimitSpec{}, proxy.ParseTrustedProxies(""))
	if !l.Allow(httptest.NewRequest(http.MethodGet, "/", nil)) {
		t.Fatal("disabled")
	}
}
