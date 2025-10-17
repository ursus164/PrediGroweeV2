package handlers

import (
	"auth/internal/models"
	"auth/internal/storage"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type SummaryHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewSummaryHandler(storage storage.Store, logger *zap.Logger) *SummaryHandler {
	return &SummaryHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *SummaryHandler) Handle(w http.ResponseWriter, _ *http.Request) {
	var summary models.AuthSummary
	summary.Users = h.storage.GetUsersCount()
	summary.ActiveUsers = h.storage.GetActiveUsersCount()
	summary.Last24hRegistered = h.storage.GetLast24hRegisteredCount()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(summary)
	if err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
