// Package middleware provides HTTP middleware for the admin service.
package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

// InternalAuth is a middleware that verifies internal API key authentication.
func InternalAuth(next http.HandlerFunc, _ *zap.Logger, apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" {
			http.Error(w, "internal api key not configured", http.StatusInternalServerError)
			return
		}
		if r.Header.Get("X-Api-Key") != apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
