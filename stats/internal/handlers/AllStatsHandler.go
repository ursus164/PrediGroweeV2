// Package handlers provides HTTP request handlers for statistics endpoints.
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

// GetAllStatsHandler handles HTTP requests for retrieving various statistics.
type GetAllStatsHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

// NewGetAllStatsHandler creates a new GetAllStatsHandler instance.
func NewGetAllStatsHandler(storage storage.Storage, logger *zap.Logger) *GetAllStatsHandler {
	return &GetAllStatsHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetResponses retrieves all quiz question responses.
func (h *GetAllStatsHandler) GetResponses(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.storage.GetAllResponses()
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	statsJSON, _ := json.Marshal(stats)
	_, err = w.Write(statsJSON)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// GetStatsForQuestion retrieves statistics for a specific question or all questions.
func (h *GetAllStatsHandler) GetStatsForQuestion(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetStatsForQuestion")
	questionID := r.PathValue("id")
	if questionID == "-" {
		stats, err := h.storage.GetStatsForAllQuestions()
		if err != nil {
			h.logger.Error("failed to get stats", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		statsJSON, _ := json.Marshal(stats)
		if _, writeErr := w.Write(statsJSON); writeErr != nil {
			h.logger.Error("failed to write response", zap.Error(writeErr))
		}
		return
	}
	questionIDInt, err := strconv.Atoi(questionID)
	if err != nil {
		h.logger.Error("failed to parse question id", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	stats, err := h.storage.GetStatsForQuestion(questionIDInt)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	statsJSON, _ := json.Marshal(stats)
	_, err = w.Write(statsJSON)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// GetActivity retrieves daily activity statistics.
func (h *GetAllStatsHandler) GetActivity(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.storage.GetActivityStats()
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	statsJSON, _ := json.Marshal(stats)
	_, err = w.Write(statsJSON)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// GetSummary retrieves a summary of all statistics.
func (h *GetAllStatsHandler) GetSummary(w http.ResponseWriter, _ *http.Request) {
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

// GetStatsGroupedBySurvey retrieves statistics grouped by survey field.
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

// DeleteResponse deletes a specific response by ID.
func (h *GetAllStatsHandler) DeleteResponse(w http.ResponseWriter, r *http.Request) {
	resID := r.PathValue("id")
	ID, err := strconv.Atoi(resID)
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
