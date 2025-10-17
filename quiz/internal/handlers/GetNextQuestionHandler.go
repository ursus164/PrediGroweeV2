package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/storage"
	"strconv"
	"time"
)

type GetNextQuestionHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewGetNextQuestionHandler(storage storage.Store, logger *zap.Logger) *GetNextQuestionHandler {
	return &GetNextQuestionHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *GetNextQuestionHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("get next question handler")

	sessionId := r.PathValue("quizSessionId")
	if sessionId == "" {
		http.Error(rw, "missing session id", http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.Atoi(sessionId)
	if err != nil {
		http.Error(rw, "invalid session id", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int)

	session, err := h.storage.GetQuizSessionByID(sessionID)
	if err != nil {
		h.logger.Error("failed to get session", zap.Error(err))
		http.Error(rw, "failed to get session", http.StatusNotFound)
		return
	}
	if session.UserID != userID {
		http.Error(rw, "failed to get session", http.StatusNotFound)
		return
	}
	if session.Status == "finished" {
		http.Error(rw, "quiz is finished", http.StatusNotFound)
		return
	}

	if session.CurrentQuestionID <= 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	question, err := h.storage.GetQuestionByID(session.CurrentQuestionID)
	if err != nil {
		http.Error(rw, "failed to get question", http.StatusNotFound)
		return
	}

	for i := range question.Case.ParameterValues {
		question.Case.ParameterValues[i].Value3 = nil
	}

	isLast := false
	if session.TestID != nil {
		for i, q := range session.GroupOrder {
			if q == session.CurrentQuestionID && i == len(session.GroupOrder)-1 {
				isLast = true
				break
			}
		}
	}

	rw.Header().Set("X-Quiz-Is-Last", strconv.FormatBool(isLast))
	rw.Header().Set("Content-Type", "application/json")

	session.QuestionRequestedTime = time.Now()
	if err := h.storage.UpdateQuizSession(session); err != nil {
		h.logger.Error("failed to update session, will result in wrong answer time", zap.Error(err))
	}

	resp := map[string]interface{}{
		"question": question,
		"is_last":  isLast,
	}
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		http.Error(rw, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
