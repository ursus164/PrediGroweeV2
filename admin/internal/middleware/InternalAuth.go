package middleware

import (
	"go.uber.org/zap"
	"net/http"
)

func InternalAuth(next http.HandlerFunc, logger *zap.Logger, apiKey string) http.HandlerFunc {
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
