package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"stats/internal/storage"
)

type SessionsAccuracyHandler struct {
	store  storage.Storage
	logger *zap.Logger
}

func NewSessionsAccuracyHandler(s storage.Storage, l *zap.Logger) *SessionsAccuracyHandler {
	return &SessionsAccuracyHandler{store: s, logger: l}
}

func (h *SessionsAccuracyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("ids")
	if q == "" {
		http.Error(w, "missing ids", http.StatusBadRequest)
		return
	}
	parts := strings.Split(q, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			http.Error(w, "bad ids", http.StatusBadRequest)
			return
		}
		ids = append(ids, n)
	}

	data, err := h.store.GetAccuracyBatch(ids)
	if err != nil {
		h.logger.Error("GetAccuracyBatch failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

