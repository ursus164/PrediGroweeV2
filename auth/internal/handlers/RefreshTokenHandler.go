package handlers

import (
	"auth/internal/auth"
	"auth/internal/storage"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type RefreshTokenHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewRefreshTokenHandler(store storage.Store, logger *zap.Logger) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		store:  store,
		logger: logger,
	}
}

func (h *RefreshTokenHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	accessToken, err := auth.GenerateAccessToken(strconv.Itoa(userID))
	if err != nil {
		h.logger.Error("Error generating access token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken, "message": "refresh successful"})
	if err != nil {
		h.logger.Error("Error encoding response", zap.Error(err))
		return
	}
}
