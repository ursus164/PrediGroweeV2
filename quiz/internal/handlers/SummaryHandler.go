package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/models"
	"quiz/internal/storage"
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
	var summary models.QuizSummary
	var err error
	summary.Questions, err = h.storage.CountQuestions()
	if err != nil {
		h.logger.Error("failed to count questions", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	summary.ActiveSurveys = 0

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(summary)
	if err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
