package handlers

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"stats/internal/models"
	"stats/internal/storage"
	"strconv"
)

type GetAllStatsHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

func NewGetAllStatsHandler(storage storage.Storage, logger *zap.Logger) *GetAllStatsHandler {
	return &GetAllStatsHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *GetAllStatsHandler) GetResponses(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.storage.GetAllResponses()
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	statsJson, _ := json.Marshal(stats)
	_, err = w.Write(statsJson)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *GetAllStatsHandler) GetStatsForQuestion(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetStatsForQuestion")
	questionId := r.PathValue("id")
	if questionId == "-" {
		stats, err := h.storage.GetStatsForAllQuestions()
		if err != nil {
			h.logger.Error("failed to get stats", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		statsJson, _ := json.Marshal(stats)
		_, err = w.Write(statsJson)
		return
	}
	questionID, err := strconv.Atoi(questionId)
	if err != nil {
		h.logger.Error("failed to parse question id", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	stats, err := h.storage.GetStatsForQuestion(questionID)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	statsJson, _ := json.Marshal(stats)
	_, err = w.Write(statsJson)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *GetAllStatsHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	stats, err := h.storage.GetActivityStats()
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	statsJson, _ := json.Marshal(stats)
	_, err = w.Write(statsJson)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *GetAllStatsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	var summary models.StatsSummary
	var err error
	summary.QuizSessions, err = h.storage.CountQuizSessions()
	if err != nil {
		h.logger.Error("failed to count quiz sessions", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	summary.TotalResponses, err = h.storage.CountAnswers()
	if err != nil {
		h.logger.Error("failed to count answers", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	summary.TotalCorrect, err = h.storage.CountCorrectAnswers()
	if err != nil {
		h.logger.Error("failed to count correct answers", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(summary)
	if err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *GetAllStatsHandler) GetStatsGroupedBySurvey(w http.ResponseWriter, r *http.Request) {
	groupBy := r.URL.Query().Get("groupBy")
	if groupBy == "" {
		http.Error(w, "groupBy parameter is required", http.StatusBadRequest)
		return
	}
	stats, err := h.storage.GetStatsGroupedBySurveyField(groupBy)
	if err != nil {
		h.logger.Error("failed to get grouped stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(stats)
	if err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *GetAllStatsHandler) DeleteResponse(w http.ResponseWriter, r *http.Request) {
	resId := r.PathValue("id")
	ID, err := strconv.Atoi(resId)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	err = h.storage.DeleteResponse(ID)
	if err != nil {
		http.Error(w, "couldn't delete response", http.StatusInternalServerError)
		return
	}
}
