// Package auth validates the shared API key (constant-time).
package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// ExtractKey returns the presented API key from headers or form field api_key.
func ExtractKey(r *http.Request) string {
	if h := strings.TrimSpace(r.Header.Get("X-API-Key")); h != "" {
		return h
	}
	if ah := strings.TrimSpace(r.Header.Get("Authorization")); ah != "" {
		const p = "Bearer "
		if len(ah) > len(p) && strings.EqualFold(ah[:len(p)], p) {
			return strings.TrimSpace(ah[len(p):])
		}
	}
	_ = r.ParseForm()
	return strings.TrimSpace(r.Form.Get("api_key"))
}

// Valid reports whether presented matches expected (constant-time when same length).
func Valid(presented, expected string) bool {
	if expected == "" {
		return false
	}
	a := []byte(presented)
	b := []byte(expected)
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}
