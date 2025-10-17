package handlers

import (
	"auth/internal/models"
	"auth/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type LogOutHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewLogOutHandler(store storage.Store, logger *zap.Logger) *LogOutHandler {
	return &LogOutHandler{
		store:  store,
		logger: logger,
	}
}

func (h *LogOutHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	err := h.store.UpdateUserSession(models.UserSession{
		UserID:     userID,
		SessionID:  "",
		Expiration: time.Now(),
	})
	if err != nil {
		h.logger.Error("Error updating session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
