package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"stats/internal/clients"
)

type contextKey string

const (
	contextKeyUserID   contextKey = "user_id"
	contextKeyUserRole contextKey = "user_role"
)

// VerifyToken is a middleware that validates authentication tokens and adds user data to context.
func VerifyToken(next http.HandlerFunc, authClient *clients.AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := ExtractAccessTokenFromRequest(r)
		log.Println("Extracted token: ", accessToken)
		if err != nil {
			log.Println("No token cookie provided")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		userData, err := authClient.VerifyAuthToken(accessToken)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			log.Println("failed to verify token: ", err)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), contextKeyUserID, userData.UserID))
		r = r.WithContext(context.WithValue(r.Context(), contextKeyUserRole, userData.Role))
		log.Println("completed token verification")
		next(w, r)
	}
}

// ExtractAccessTokenFromRequest extracts the access token from the request headers or cookies.
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
