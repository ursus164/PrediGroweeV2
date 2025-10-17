package handlers

import (
	"admin/clients"
	"admin/internal/models"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type SummaryHandler struct {
	logger      *zap.Logger
	authClient  clients.AuthClient
	statsClient clients.StatsClient
	quizClient  clients.QuizClient
}

func NewSummaryHandler(logger *zap.Logger, authClient clients.AuthClient, statsClient clients.StatsClient, quizClient clients.QuizClient) *SummaryHandler {
	return &SummaryHandler{
		logger:      logger,
		authClient:  authClient,
		statsClient: statsClient,
		quizClient:  quizClient,
	}
}

func (h *SummaryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	wg := sync.WaitGroup{}
	wg.Add(3)
	var errs []error
	var summary models.Summary
	go func() {
		defer wg.Done()
		quizSummary, err := h.quizClient.GetSummary()
		if err != nil {
			errs = append(errs, err)
		}
		summary.QuizSummary = quizSummary
	}()
	go func() {
		defer wg.Done()
		statsSummary, err := h.statsClient.GetSummary()
		if err != nil {
			errs = append(errs, err)
		}
		summary.StatsSummary = statsSummary
	}()
	go func() {
		defer wg.Done()
		authSummary, err := h.authClient.GetSummary()
		if err != nil {
			errs = append(errs, err)
		}
		summary.AuthSummary = authSummary
	}()
	wg.Wait()
	if len(errs) != 0 {
		err := fmt.Errorf(fmt.Sprintf("failed to get summary: %v", errs))
		h.logger.Error("failed to get summary", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	summaryJson, _ := json.Marshal(summary)
	_, err := w.Write(summaryJson)
	if err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
