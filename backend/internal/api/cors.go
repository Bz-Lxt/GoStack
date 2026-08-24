package api

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

func allowOrigin(origin, host string, whitelist []string) bool {
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if host != "" && strings.EqualFold(u.Host, host) {
		return true
	}
	h, _, err := net.SplitHostPort(u.Host)
	if err == nil && (h == "127.0.0.1" || h == "localhost" || h == "::1") {
		return true
	}
	for _, w := range whitelist {
		if strings.EqualFold(strings.TrimSpace(w), origin) {
			return true
		}
	}
	return false
}

func withCORS(next http.Handler, whitelist []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowOrigin(origin, r.Host, whitelist) && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
