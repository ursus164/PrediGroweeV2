// Package middleware contains HTTP middleware utilities for the Images service.
package middleware

import (
	"fmt"
	"images/internal/clients"
	"log"
	"net/http"
)

// VerifyToken ensures that a valid access token is present before executing the next handler.
func VerifyToken(next http.HandlerFunc, authClient *clients.AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := ExtractAccessTokenFromRequest(r)
		log.Println("Extracted token: ", accessToken)
		if err != nil {
			log.Println("No token cookie provided")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		err = authClient.VerifyAuthToken(accessToken)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			log.Println("failed to verify token: ", err)
			return
		}
		log.Println("completed token verification")
		next(w, r)
	}
}
// ExtractAccessTokenFromRequest extracts the access token from Authorization header or cookie.
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
