package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValid(t *testing.T) {
	if !Valid("abc", "abc") {
		t.Fatal("same key")
	}
	if Valid("abc", "abd") {
		t.Fatal("different")
	}
	if Valid("ab", "abc") {
		t.Fatal("length")
	}
}

func TestExtractKey(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("X-API-Key", "k1")
	if ExtractKey(r) != "k1" {
		t.Fatal("x-api-key")
	}
	r2 := httptest.NewRequest(http.MethodPost, "/", nil)
	r2.Header.Set("Authorization", "Bearer k2")
	if ExtractKey(r2) != "k2" {
		t.Fatal("bearer")
	}
	r3 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("api_key=k3"))
	r3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if ExtractKey(r3) != "k3" {
		t.Fatal("form")
	}
}
