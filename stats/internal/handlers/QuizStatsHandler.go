package handlers

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"stats/internal/models"
	"stats/internal/storage"
	"strconv"
)

// QuizStatsHandler handles HTTP requests for quiz statistics.
type QuizStatsHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

// NewQuizStatsHandler creates a new QuizStatsHandler instance.
func NewQuizStatsHandler(store storage.Storage, logger *zap.Logger) *QuizStatsHandler {
	return &QuizStatsHandler{
		storage: store,
		logger:  logger,
	}
}

// GetStats retrieves statistics for a specific quiz session.
func (h *QuizStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	quizID := r.PathValue("quizSessionId")
	if quizID == "" {
		http.Error(w, "missing quiz id", http.StatusBadRequest)
		return
	}
	quizSessionID, err := strconv.Atoi(quizID)
	if err != nil {
		http.Error(w, "invalid quiz id", http.StatusBadRequest)
		return
	}
	session, err := h.storage.GetQuizSessionByID(quizSessionID)
	if err != nil {
		http.Error(w, "failed to get session", http.StatusNotFound)
		return
	}
	if session.UserID != userID {
		http.Error(w, "failed to get session", http.StatusNotFound)
		return
	}
	stats, err := h.storage.GetUserQuizStats(quizSessionID)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := stats.ToJSON(w); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}

// SaveSession saves a new quiz session.
func (h *QuizStatsHandler) SaveSession(w http.ResponseWriter, r *http.Request) {
	var sessionData models.QuizSession
	err := sessionData.FromJSON(r.Body)
	if err != nil {
		http.Error(w, "failed to decode session", http.StatusBadRequest)
		return
	}
	err = h.storage.SaveSession(&sessionData)
	if err != nil {
		http.Error(w, "failed to save session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// FinishSession marks a quiz session as finished.
func (h *QuizStatsHandler) FinishSession(w http.ResponseWriter, r *http.Request) {
	quizID := r.PathValue("quizSessionId")
	if quizID == "" {
		http.Error(w, "missing quiz id", http.StatusBadRequest)
		return
	}
	quizSessionID, err := strconv.Atoi(quizID)
	if err != nil {
		http.Error(w, "invalid quiz id", http.StatusBadRequest)
		return
	}
	err = h.storage.FinishQuizSession(quizSessionID)
	if err != nil {
		http.Error(w, "failed to finish session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SaveResponse saves a question response for a quiz session.
func (h *QuizStatsHandler) SaveResponse(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.PathValue("quizSessionId")
	if sessionIDStr == "" {
		http.Error(w, "missing session id", http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		h.logger.Error("failed to parse session id", zap.Error(err))
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	h.logger.Info("SaveResponseHandler.GetResponses")
	var response models.QuestionResponse
	err = response.FromJSON(r.Body)
	h.logger.Info(fmt.Sprintf("response to save: %+v", response))
	if err != nil {
		h.logger.Error("failed to decode response", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	session, err := h.storage.GetQuizSessionByID(sessionID)
	if errors.Is(err, storage.ErrSessionNotFound) {
		err = h.storage.SaveSession(&models.QuizSession{
			SessionID: sessionID,
		})
		if err != nil {
			h.logger.Error("failed to save session", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		h.logger.Error("failed to get session", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if session.FinishTime != nil {
		h.logger.Error("session already finished")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// todo: check if response already exists
	if err := h.storage.SaveResponse(sessionID, &response); err != nil {
		h.logger.Error("failed to save response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
