package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"
	"quiz/internal/storage"
)

type ListActiveSessionsHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewListActiveSessionsHandler(s storage.Store, l *zap.Logger) *ListActiveSessionsHandler {
	return &ListActiveSessionsHandler{storage: s, logger: l}
}

func (h *ListActiveSessionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cutoff, _ := strconv.Atoi(q.Get("cutoff"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	data, err := h.storage.ListActiveSessions(cutoff, limit)
	if err != nil {
		h.logger.Error("ListActiveSessions failed", zap.Error(err))
		http.Error(w, "Failed to list active sessions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

