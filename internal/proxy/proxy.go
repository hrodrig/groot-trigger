// Package proxy resolves client IPs behind optional trusted reverse proxies.
package proxy

import (
	"net"
	"net/http"
	"strings"
)

// TrustedProxies is the set of TCP peers allowed to supply X-Forwarded-For / X-Real-IP.
type TrustedProxies struct {
	nets []*net.IPNet
}

// ParseTrustedProxies parses comma-separated IPs/CIDRs.
func ParseTrustedProxies(s string) *TrustedProxies {
	tp := &TrustedProxies{}
	for _, raw := range strings.Split(s, ",") {
		cidr := strings.TrimSpace(raw)
		if cidr == "" {
			continue
		}
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		tp.nets = append(tp.nets, n)
	}
	return tp
}

// Empty reports whether no proxies are trusted (forwarded headers ignored).
func (t *TrustedProxies) Empty() bool {
	return t == nil || len(t.nets) == 0
}

// ContainsIP reports whether ip is inside any trusted network.
func (t *TrustedProxies) ContainsIP(ip net.IP) bool {
	if t == nil || ip == nil {
		return false
	}
	for _, n := range t.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ClientIP returns the client IP for rate limiting and access logs.
func ClientIP(r *http.Request, trusted *TrustedProxies) string {
	peer := peerIP(r.RemoteAddr)
	peerParsed := net.ParseIP(peer)
	if trusted.Empty() || peerParsed == nil || !trusted.ContainsIP(peerParsed) {
		return peer
	}
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		for _, hop := range strings.Split(xff, ",") {
			ip := strings.TrimSpace(hop)
			if ip == "" {
				continue
			}
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
		return peer
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}
	return peer
}

func peerIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
