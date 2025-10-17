package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jordan-wright/email"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func ValidateJWT(tokenString string) (token *jwt.Token, err error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func GenerateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(10 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
func ExtractAccessTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) == 2 {
			return splitToken[1], nil
		}
	}
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("no valid access token found")
}
func ExtractSessionIdFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", fmt.Errorf("empty session id")
	}
	return cookie.Value, nil
}
func GenerateSessionID(length int) (string, error) {
	bytes := make([]byte, length)
	// Fill the byte slice with random bytes
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	// Convert the byte slice to a hex string
	return hex.EncodeToString(bytes), nil
}

// todo: use this function to set cookies in handlers
func SetCookie(w http.ResponseWriter, name, value string) {
	isProd := os.Getenv("ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Path:     "/",
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
	})
}

func SendVerificationEmail(to, token string) error {
	e := email.NewEmail()
	e.From = "PrediGrowee <noreply@predigrowee.agh.edu.pl>"
	e.To = []string{to}
	e.Subject = "Verify your email"
	e.HTML = []byte(fmt.Sprintf(`
        <h1>Verify your email</h1>
        <p>Click <a href="https://predigrowee.agh.edu.pl/api/auth/verify-email?token=%s">here</a> to verify your email.</p>
    `, token))

	// For Gmail:
	return e.Send("smtp.gmail.com:587", smtp.PlainAuth(
		"",
		os.Getenv("GMAIL_USER"),
		os.Getenv("GMAIL_PASSWORD"),
		"smtp.gmail.com",
	))

}

func GenerateVerificationToken(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": id,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func SendPasswordResetEmail(to, token string) error {
	e := email.NewEmail()
	e.From = "PrediGrowee <noreply@predigrowee.agh.edu.pl>"
	e.To = []string{to}
	e.Subject = "Reset your password"
	e.HTML = []byte(fmt.Sprintf(`
        <h1>Reset your password</h1>
        <p>Click <a href="https://predigrowee.agh.edu.pl/reset-password?token=%s">here</a> to reset your password.</p>
    `, token))

	return e.Send("smtp.gmail.com:587", smtp.PlainAuth(
		"",
		os.Getenv("GMAIL_USER"),
		os.Getenv("GMAIL_PASSWORD"),
		"smtp.gmail.com",
	))
}

func SendApprovedEmail(to string) error {
	e := email.NewEmail()
	e.From = "PrediGrowee <noreply@predigrowee.agh.edu.pl>"
	e.To = []string{to}
	e.Subject = "Your account was verified by administrator"
	e.HTML = []byte(`
        <p>Administrator has approved your account</p>
        <p>You can start your quiz right now: <a href="https://predigrowee.agh.edu.pl/login">Log in</a></p>
    `)

	return e.Send("smtp.gmail.com:587", smtp.PlainAuth(
		"",
		os.Getenv("GMAIL_USER"),
		os.Getenv("GMAIL_PASSWORD"),
		"smtp.gmail.com",
	))
}
