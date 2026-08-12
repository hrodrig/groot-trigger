// Package ratelimit provides in-process token-bucket limits.
package ratelimit

import (
	"net/http"
	"sync"

	"github.com/hrodrig/groot-trigger/internal/config"
	"github.com/hrodrig/groot-trigger/internal/proxy"
	"golang.org/x/time/rate"
)

// Limiter gates POST collect by client IP and optional global cap.
type Limiter struct {
	perIP   *bucketMap
	global  *rate.Limiter
	trusted *proxy.TrustedProxies
}

type bucketMap struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

// New builds a limiter. Zero LimitSpec disables that dimension.
func New(post, global config.LimitSpec, trusted *proxy.TrustedProxies) *Limiter {
	l := &Limiter{trusted: trusted}
	if post.Requests > 0 && post.Window > 0 {
		l.perIP = &bucketMap{
			limiters: make(map[string]*rate.Limiter),
			r:        rate.Limit(float64(post.Requests) / post.Window.Seconds()),
			burst:    max(1, post.Requests),
		}
	}
	if global.Requests > 0 && global.Window > 0 {
		r := rate.Limit(float64(global.Requests) / global.Window.Seconds())
		l.global = rate.NewLimiter(r, max(1, global.Requests))
	}
	return l
}

// Allow reports whether the request may proceed.
func (l *Limiter) Allow(r *http.Request) bool {
	if l == nil {
		return true
	}
	if l.global != nil && !l.global.Allow() {
		return false
	}
	if l.perIP == nil {
		return true
	}
	ip := proxy.ClientIP(r, l.trusted)
	return l.perIP.allow(ip)
}

func (m *bucketMap) allow(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	lim, ok := m.limiters[key]
	if !ok {
		lim = rate.NewLimiter(m.r, m.burst)
		m.limiters[key] = lim
	}
	return lim.Allow()
}
