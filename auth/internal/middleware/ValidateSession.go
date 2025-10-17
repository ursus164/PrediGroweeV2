package middleware

import (
	"auth/internal/auth"
	"auth/internal/storage"
	"context"
	"log"
	"net/http"
	"time"
)

func ValidateSession(next http.HandlerFunc, storage storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := auth.ExtractSessionIdFromCookie(r)
		log.Println(sessionID)
		if err != nil {
			log.Println("Failed to extract session ID", err)
			http.Error(w, "Invalid session ID", http.StatusUnauthorized)
			return
		}
		session, err := storage.GetUserSessionBySessionID(sessionID)
		if err != nil {
			log.Println("Failed to get session from storage", err)
			http.Error(w, "Invalid session id", http.StatusUnauthorized)
			return
		}
		if session.Expiration.Before(time.Now()) {
			log.Println("Session expired")
			http.Error(w, "Session expired. Please log in", http.StatusUnauthorized)
			return
		}
		user, err := storage.GetUserById(session.UserID, false)
		if err != nil {
			log.Printf("Failed to get user from storage: %v", err)
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		log.Printf("User found: %+v", user)
		newCtx := context.WithValue(r.Context(), "user_id", session.UserID)
		newCtx = context.WithValue(newCtx, "user_role", user.Role)
		next(w, r.WithContext(newCtx))
	}
}
