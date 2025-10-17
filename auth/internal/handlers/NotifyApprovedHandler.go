package handlers

import (
	"auth/internal/auth"
	"auth/internal/storage"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type NotifyApprovedHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewNotifyApprovedHandler(s storage.Store, l *zap.Logger) *NotifyApprovedHandler {
	return &NotifyApprovedHandler{store: s, logger: l}
}

type notifyApprovedReq struct {
	UserID *int    `json:"user_id,omitempty"`
	Email  *string `json:"email,omitempty"`
}

func (h *NotifyApprovedHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req notifyApprovedReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var emailAddr string
	switch {
	case req.Email != nil && *req.Email != "":
		emailAddr = *req.Email
	case req.UserID != nil:
		u, err := h.store.GetUserById(*req.UserID, false)
		if err != nil || u == nil || u.Email == "" {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		emailAddr = u.Email
	default:
		http.Error(w, "email or user_id required", http.StatusBadRequest)
		return
	}

	if err := auth.SendApprovedEmail(emailAddr); err != nil {
		h.logger.Error("failed to send approved email", zap.Error(err))
		http.Error(w, "failed to send", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

