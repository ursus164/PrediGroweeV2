package middleware

import (
	"auth/internal/auth"
	"auth/internal/storage"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strconv"
)

func ValidateAccessToken(next http.HandlerFunc, storage storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Validating access token")

		tokenString, err := auth.ExtractAccessTokenFromRequest(r)
		if err != nil {
			log.Printf("Failed to extract token: %v", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		log.Printf("Extracted token: %s", tokenString)

		token, err := auth.ValidateJWT(tokenString)
		if err != nil {
			log.Printf("Failed to validate JWT: %v", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			log.Println("Token is invalid")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		log.Println("JWT validated successfully")

		tokenClaims := token.Claims.(jwt.MapClaims)
		userID, err := strconv.Atoi(tokenClaims["sub"].(string))
		if err != nil {
			log.Printf("Failed to convert user ID: %v", err)
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		log.Printf("User ID from token: %d", userID)

		user, err := storage.GetUserById(userID, false)
		if err != nil {
			log.Printf("Failed to get user from storage: %v", err)
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		log.Printf("User found: %+v", user)

		// Add user information to the request context
		ctx := context.WithValue(r.Context(), "user_id", user.ID)
		ctx = context.WithValue(ctx, "user_role", user.Role)
		r = r.WithContext(ctx)

		log.Println("Access token validated successfully")
		next(w, r)
	}
}
