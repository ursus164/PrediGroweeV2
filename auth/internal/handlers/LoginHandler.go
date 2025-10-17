package handlers

import (
	"auth/internal/auth"
	"auth/internal/models"
	"auth/internal/storage"
	"encoding/json"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LoginHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewLoginHandler(store storage.Store, logger *zap.Logger) *LoginHandler {
	return &LoginHandler{
		logger: logger,
		store:  store,
	}
}

func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var userPayload models.LoginUserPayload
	if err := userPayload.FromJSON(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userPayload.Email = strings.ToLower(userPayload.Email)
	dbUser, err := h.store.GetUserByEmail(userPayload.Email)
	if err != nil {
		h.logger.Error("Error getting user by email", zap.Error(err))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if !dbUser.Verified {
		http.Error(w, "User not verified", http.StatusUnauthorized)
		return
	}
	//userSession, err := h.store.GetUserSession(dbUser.ID)
	//if err == nil {
	//	if userSession.Expiration.After(time.Now()) {
	//		http.Error(w, "User already logged in", http.StatusConflict)
	//		return
	//	}
	//}
	if err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(userPayload.Password)); err != nil {
		h.logger.Error("Error comparing passwords", zap.Error(err))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	sessionId, err := auth.GenerateSessionID(64)
	if err != nil {
		h.logger.Error("Error generating session id", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = h.store.SaveUserSession(models.UserSession{
		UserID:     dbUser.ID,
		SessionID:  sessionId,
		Expiration: time.Now().Add(7 * 24 * time.Hour),
	})
	accessToken, err := auth.GenerateAccessToken(strconv.Itoa(dbUser.ID))
	if err != nil {
		h.logger.Error("Error generating access token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	auth.SetCookie(w, "session_id", sessionId)
	w.Header().Set("Content-Type", "application/json")
	data := map[string]interface{}{"user_id": dbUser.ID, "role": dbUser.Role, "access_token": accessToken}
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		h.logger.Error("Error encoding response", zap.Error(err))
		return
	}
}
