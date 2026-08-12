package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPIgnoresForwardedWhenUntrusted(t *testing.T) {
	t.Parallel()
	tp := ParseTrustedProxies("")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:1234"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	if got := ClientIP(r, tp); got != "10.0.0.5" {
		t.Fatalf("got %q", got)
	}
}

func TestClientIPUsesXFFWhenTrusted(t *testing.T) {
	t.Parallel()
	tp := ParseTrustedProxies("10.0.0.0/8")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:1234"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	if got := ClientIP(r, tp); got != "1.2.3.4" {
		t.Fatalf("got %q", got)
	}
}
