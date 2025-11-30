package middleware

import (
	"admin/clients"
	"admin/internal/models"
	"context"
	"fmt"
	"log"
	"net/http"
)

// ContextKey is a typed key for context values to avoid collisions.
type ContextKey string

const (
	// UserIDKey is the context key for user ID.
	UserIDKey ContextKey = "user_id"
	// UserRoleKey is the context key for user role.
	UserRoleKey ContextKey = "user_role"
)

// VerifyAdmin is a middleware that verifies admin or teacher authentication.
func VerifyAdmin(next http.HandlerFunc, authClient clients.AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := ExtractAccessTokenFromRequest(r)
		log.Println("Extracted token: ", accessToken)
		if err != nil {
			log.Println("No access token provided")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		userData, err := authClient.VerifyAuthToken(accessToken)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			log.Println("failed to verify token: ", err)
			return
		}
		if userData.Role != models.RoleAdmin {
			if userData.Role != models.RoleTeacher {
				http.Error(w, "forbidden", http.StatusForbidden)
				log.Println("not admin user attempted admin action")
				return
			}
			if r.Method != http.MethodGet {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		r = r.WithContext(context.WithValue(r.Context(), UserIDKey, userData.UserID))
		r = r.WithContext(context.WithValue(r.Context(), UserRoleKey, userData.Role))
		log.Println("completed token verification")
		next(w, r)
	}
}

// ExtractAccessTokenFromRequest extracts the access token from the request.
func ExtractAccessTokenFromRequest(r *http.Request) (string, error) {
	authHeaderValue := r.Header.Get("Authorization")
	if authHeaderValue != "" {
		return authHeaderValue, nil
	}
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", fmt.Errorf("empty access token")
	}
	return cookie.Value, nil
}
