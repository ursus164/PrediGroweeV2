package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"
	"stats/internal/storage"
)

type LeaderboardHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

func NewLeaderboardHandler(s storage.Storage, l *zap.Logger) *LeaderboardHandler {
	return &LeaderboardHandler{storage: s, logger: l}
}

func (h *LeaderboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 100
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}
	minAnswers := 10
	if v := q.Get("minAnswers"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			minAnswers = n
		}
	}

	rows, err := h.storage.GetLeaderboard(minAnswers, limit)
	if err != nil {
		h.logger.Error("get leaderboard failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}

