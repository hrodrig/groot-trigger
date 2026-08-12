package proxy

import (
	"net"
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

func TestClientIPUsesXRealIP(t *testing.T) {
	t.Parallel()
	tp := ParseTrustedProxies("10.0.0.5")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:9"
	r.Header.Set("X-Real-IP", "203.0.113.9")
	if got := ClientIP(r, tp); got != "203.0.113.9" {
		t.Fatalf("got %q", got)
	}
}

func TestParseTrustedProxiesSkipsBad(t *testing.T) {
	t.Parallel()
	tp := ParseTrustedProxies("not-a-cidr, , 192.0.2.1")
	if tp.Empty() {
		t.Fatal("expected one net")
	}
	if !tp.ContainsIP(mustIP("192.0.2.1")) {
		t.Fatal("missing")
	}
	if tp.ContainsIP(nil) {
		t.Fatal("nil")
	}
}

func TestPeerIPBare(t *testing.T) {
	t.Parallel()
	if peerIP("10.1.2.3") != "10.1.2.3" {
		t.Fatal("bare")
	}
}

func mustIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		panic(s)
	}
	return ip
}
