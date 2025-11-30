// Package handlers provides HTTP request handlers for the admin service.
package handlers

import (
	"admin/clients"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// AllStatsHandler handles statistics-related HTTP requests.
type AllStatsHandler struct {
	logger      *zap.Logger
	statsClient clients.StatsClient
}

// NewAllStatsHandler creates a new AllStatsHandler instance.
func NewAllStatsHandler(logger *zap.Logger, statsClient clients.StatsClient) *AllStatsHandler {
	return &AllStatsHandler{
		logger:      logger,
		statsClient: statsClient,
	}
}

// GetAllResponses retrieves all question responses.
func (h *AllStatsHandler) GetAllResponses(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.statsClient.GetAllResponses()
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

// GetStatsForQuestion retrieves statistics for a specific question.
func (h *AllStatsHandler) GetStatsForQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("questionId")
	stats, err := h.statsClient.GetStatsForQuestion(questionID)
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

// GetStatsForAllQuestions retrieves statistics for all questions.
func (h *AllStatsHandler) GetStatsForAllQuestions(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.statsClient.GetStatsForAllQuestions()
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

// GetActivityStats retrieves activity statistics.
func (h *AllStatsHandler) GetActivityStats(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.statsClient.GetActivityStats()
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

// GetStatsGroupedBySurvey retrieves statistics grouped by survey parameter.
func (h *AllStatsHandler) GetStatsGroupedBySurvey(w http.ResponseWriter, r *http.Request) {
	groupBy := r.URL.Query().Get("groupBy")
	if groupBy == "" {
		http.Error(w, "groupBy parameter is required", http.StatusBadRequest)
		return
	}

	stats, err := h.statsClient.GetStatsGroupedBySurvey(groupBy)
	if err != nil {
		h.logger.Error("failed to get grouped stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// DeleteResponse deletes a specific response.
func (h *AllStatsHandler) DeleteResponse(w http.ResponseWriter, r *http.Request) {
	responseID := r.PathValue("id")
	err := h.statsClient.DeleteResponse(responseID)
	if err != nil {
		h.logger.Error("failed to delete response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetStatsForUsers retrieves quiz statistics for all users.
func (h *AllStatsHandler) GetStatsForUsers(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.statsClient.GetAllUsersStats()
	if err != nil {
		h.logger.Error("failed to get user stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err = json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}
