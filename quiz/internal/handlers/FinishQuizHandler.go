package handlers

import (
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/clients"
	"quiz/internal/models"
	"quiz/internal/storage"
	"strconv"
	"time"
)

type FinishQuizHandler struct {
	storage     storage.Store
	logger      *zap.Logger
	statsClient *clients.StatsClient
}

func NewFinishQuizHandler(store storage.Store, logger *zap.Logger, client *clients.StatsClient) *FinishQuizHandler {
	return &FinishQuizHandler{
		storage:     store,
		logger:      logger,
		statsClient: client,
	}
}
func (h *FinishQuizHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	quizSessionIdString := r.PathValue("quizSessionId")
	if quizSessionIdString == "" {
		h.logger.Info("no quiz session id provided")
		http.Error(rw, "invalid quiz session id", http.StatusBadRequest)
		return
	}
	quizSessionID, err := strconv.Atoi(quizSessionIdString)
	if err != nil {
		h.logger.Info("invalid quiz session id")
		http.Error(rw, "invalid quiz session id", http.StatusBadRequest)
		return
	}
	session, err := h.storage.GetQuizSessionByID(quizSessionID)
	if err != nil {
		h.logger.Error("failed to get quiz session from db", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
	session.Status = models.QuizStatusFinished
	finishTime := time.Now()
	session.FinishedAt = &finishTime
	err = h.storage.UpdateQuizSession(session)
	if err != nil {
		h.logger.Error("failed to update quiz session", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
	}
	err = h.statsClient.FinishSession(quizSessionID)
	if err != nil {
		h.logger.Error("failed to finish stats quiz session", zap.Error(err))
	}
	rw.WriteHeader(http.StatusOK)
}
