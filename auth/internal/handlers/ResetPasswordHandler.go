package handlers

import (
	"auth/internal/auth"
	"auth/internal/storage"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
)

type ResetPasswordHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewResetPasswordHandler(store storage.Store, logger *zap.Logger) *ResetPasswordHandler {
	return &ResetPasswordHandler{storage: store, logger: logger}
}

func (h *ResetPasswordHandler) RequestReset(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUserByEmail(payload.Email)
	if err != nil {
		// Return 200 even if email doesn't exist for security
		w.WriteHeader(http.StatusOK)
		return
	}

	token, err := auth.GenerateVerificationToken(strconv.Itoa(user.ID))
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = auth.SendPasswordResetEmail(user.Email, token)
	if err != nil {
		h.logger.Error("failed to send email", zap.Error(err))
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ResetPasswordHandler) Reset(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := auth.ValidateJWT(payload.Token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(token.Claims.(jwt.MapClaims)["sub"].(string))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("failed to hash password", zap.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	err = h.storage.UpdateUserPassword(userID, string(hashedPassword))
	if err != nil {
		h.logger.Error("failed to update password", zap.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func (h *ResetPasswordHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if _, err := auth.ValidateJWT(token); err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
