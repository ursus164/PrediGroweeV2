package handlers

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"stats/internal/models"
	"stats/internal/storage"
	"strconv"
)

type QuizStatsHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

func NewQuizStatsHandler(store storage.Storage, logger *zap.Logger) *QuizStatsHandler {
	return &QuizStatsHandler{
		storage: store,
		logger:  logger,
	}
}
func (h *QuizStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	quizId := r.PathValue("quizSessionId")
	if quizId == "" {
		http.Error(w, "missing quiz id", http.StatusBadRequest)
		return
	}
	quizSessionID, err := strconv.Atoi(quizId)
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
	err = stats.ToJSON(w)
}

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

func (h *QuizStatsHandler) FinishSession(w http.ResponseWriter, r *http.Request) {
	quizId := r.PathValue("quizSessionId")
	if quizId == "" {
		http.Error(w, "missing quiz id", http.StatusBadRequest)
		return
	}
	quizSessionID, err := strconv.Atoi(quizId)
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

func (h *QuizStatsHandler) SaveResponse(w http.ResponseWriter, r *http.Request) {
	sessionId := r.PathValue("quizSessionId")
	if sessionId == "" {
		http.Error(w, "missing session id", http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.Atoi(sessionId)
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
	if err == storage.ErrSessionNotFound {
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
	err = h.storage.SaveResponse(sessionID, &response)
	w.WriteHeader(http.StatusOK)
}
