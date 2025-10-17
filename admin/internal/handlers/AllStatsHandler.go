package handlers

import (
	"admin/clients"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type AllStatsHandler struct {
	logger      *zap.Logger
	statsClient clients.StatsClient
}

func NewAllStatsHandler(logger *zap.Logger, statsClient clients.StatsClient) *AllStatsHandler {
	return &AllStatsHandler{
		logger:      logger,
		statsClient: statsClient,
	}
}
func (h *AllStatsHandler) GetAllResponses(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.statsClient.GetAllResponses()
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

func (h *AllStatsHandler) GetStatsForQuestion(w http.ResponseWriter, r *http.Request) {
	questionId := r.PathValue("questionId")
	stats, err := h.statsClient.GetStatsForQuestion(questionId)
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

func (h *AllStatsHandler) GetStatsForAllQuestions(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.statsClient.GetStatsForAllQuestions()
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

func (h *AllStatsHandler) GetActivityStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsClient.GetActivityStats()
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
	err = json.NewEncoder(w).Encode(stats)
	if err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *AllStatsHandler) DeleteResponse(w http.ResponseWriter, r *http.Request) {
	responseId := r.PathValue("id")
	err := h.statsClient.DeleteResponse(responseId)
	if err != nil {
		h.logger.Error("failed to delete response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AllStatsHandler) GetStatsForUsers(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.statsClient.GetAllUsersStats()
	if err != nil {
		h.logger.Error("failed to get user stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
}
