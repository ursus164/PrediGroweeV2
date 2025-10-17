package middleware

import (
	"go.uber.org/zap"
	"net/http"
)

func InternalAuth(next http.HandlerFunc, logger *zap.Logger, validAPIKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("InternalAuth middleware")

		apiKey := r.Header.Get("X-Api-Key")
		if apiKey != validAPIKey {
			logger.Warn("Invalid API key")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
