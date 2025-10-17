package handlers

import (
	"encoding/json"
	"net/http"

	"admin/clients"

	"go.uber.org/zap"
)

type TestsProgressHandler struct {
	quiz   clients.QuizClient
	stats  clients.StatsClient
	logger *zap.Logger
}

func NewTestsProgressHandler(q clients.QuizClient, s clients.StatsClient, l *zap.Logger) *TestsProgressHandler {
	return &TestsProgressHandler{quiz: q, stats: s, logger: l}
}

func (h *TestsProgressHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	sessions, err := h.quiz.ListSessionsByTestCode(code)
	if err != nil {
		h.logger.Error("ListSessionsByTestCode failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	if len(sessions) == 0 {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]any{})
		return
	}

	ids := make([]int, 0, len(sessions))
	for _, s := range sessions {
		ids = append(ids, s.ID)
	}
	accRows, err := h.stats.GetSessionsAccuracy(ids)
	if err != nil {
		h.logger.Error("GetSessionsAccuracy failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	accMap := make(map[int]struct {
		Correct  int
		Total    int
		Accuracy float64
	}, len(accRows))
	for _, a := range accRows {
		accMap[a.SessionID] = struct {
			Correct  int
			Total    int
			Accuracy float64
		}{Correct: a.Correct, Total: a.Total, Accuracy: a.Accuracy}
	}

	type row struct {
		SessionID      int      `json:"session_id"`
		UserID         int      `json:"user_id"`
		Status         string   `json:"status"`
		Mode           string   `json:"mode"`
		TestCode       string   `json:"test_code"`
		CreatedAt      string   `json:"created_at"`
		FinishedAt     *string  `json:"finished_at,omitempty"`
		Accuracy       *float64 `json:"accuracy,omitempty"`
		CorrectAnswers *int     `json:"correct_answers,omitempty"`
		TotalAnswers   *int     `json:"total_answers,omitempty"`
	}

	out := make([]row, 0, len(sessions))
	for _, s := range sessions {
		var codeStr string
		if s.TestCode != nil {
			codeStr = *s.TestCode
		}
		rw := row{
			SessionID:  s.ID,
			UserID:     s.UserID,
			Status:     s.Status,
			Mode:       s.Mode,
			TestCode:   codeStr,
			CreatedAt:  s.CreatedAt,
			FinishedAt: s.FinishedAt,
		}
		if v, ok := accMap[s.ID]; ok {
			acc := v.Accuracy
			cor := v.Correct
			tot := v.Total
			rw.Accuracy = &acc
			rw.CorrectAnswers = &cor
			rw.TotalAnswers = &tot
		}
		out = append(out, rw)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

