package handlers

import (
	"encoding/json"
	"net/http"
	"quiz/internal/storage"
    "time"
	"go.uber.org/zap"
)

type TestSessionsHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewTestSessionsHandler(s storage.Store, l *zap.Logger) *TestSessionsHandler {
	return &TestSessionsHandler{store: s, logger: l}
}

func (h *TestSessionsHandler) ListByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	t, err := h.store.GetTestByCode(code)
	if err != nil {
		h.logger.Error("GetTestByCode failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	sessions, err := h.store.ListSessionsByTestID(t.ID)
	if err != nil {
		h.logger.Error("ListSessionsByTestID failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	type row struct {
		ID              int     `json:"id"`
		UserID          int     `json:"user_id"`
		Status          string  `json:"status"`
		Mode            string  `json:"mode"`
		CurrentQuestion int     `json:"current_question"`
		CurrentGroup    int     `json:"current_group"`
		CreatedAt       string  `json:"created_at"`
		UpdatedAt       string  `json:"updated_at"`
		FinishedAt      *string `json:"finished_at,omitempty"`
		TestID          *int    `json:"test_id,omitempty"`
		TestCode        *string `json:"test_code,omitempty"`
	}

	out := make([]row, 0, len(sessions))
	for _, s := range sessions {
		var finishedAt *string
		if s.FinishedAt != nil {
			str := s.FinishedAt.Format(time.RFC3339)
			finishedAt = &str
		}
		r := row{
			ID:              s.ID,
			UserID:          s.UserID,
			Status:          s.Status,
			Mode:            s.Mode,
			CurrentQuestion: s.CurrentQuestionID,
			CurrentGroup:    s.CurrentGroup,
			CreatedAt:       s.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       s.UpdatedAt.Format(time.RFC3339),
			FinishedAt:      finishedAt,
			TestID:          s.TestID,
			TestCode:        s.TestCode,
		}
		out = append(out, r)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

