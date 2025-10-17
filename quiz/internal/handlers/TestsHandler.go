package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"quiz/internal/models"
	"quiz/internal/storage"
)

type TestsHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewTestsHandler(store storage.Store, logger *zap.Logger) *TestsHandler {
	return &TestsHandler{store: store, logger: logger}
}

type createTestReq struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	QuestionIDs []int  `json:"question_ids"`
}

var codeRe = regexp.MustCompile(`^[A-Z0-9-]{4,24}$`)

func (h *TestsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	var req createTestReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	req.Code = strings.TrimSpace(strings.ToUpper(req.Code))
	req.Name = strings.TrimSpace(req.Name)

	if req.Code == "" || !codeRe.MatchString(req.Code) {
		http.Error(w, "invalid code format", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(req.QuestionIDs) == 0 {
		http.Error(w, "question_ids cannot be empty", http.StatusBadRequest)
		return
	}

	t := models.Test{
		Code:      req.Code,
		Name:      req.Name,
		CreatedBy: userID,
	}

	created, err := h.store.CreateTest(t, req.QuestionIDs)
	if err != nil {
		// duplicate code
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			http.Error(w, "test code already exists", http.StatusConflict)
			return
		}
		h.logger.Error("create test failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (h *TestsHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	tests, err := h.store.ListTestsByOwner(userID)
	if err != nil {
		h.logger.Error("list tests failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tests)
}

type sessionRow struct {
	SessionID int        `json:"session_id"`
	UserID    int        `json:"user_id"`
	Mode      string     `json:"mode"`
	Status    string     `json:"status"`
	CreatedAt *time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func (h *TestsHandler) ProgressByCode(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(r.PathValue("code")))
	if code == "" {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}

	t, err := h.store.GetTestByCode(code)
	if err != nil {
		h.logger.Error("get test by code failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	sessions, err := h.store.ListSessionsByTestID(t.ID)
	if err != nil && err != sql.ErrNoRows {
		h.logger.Error("list sessions by test id failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]sessionRow, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, sessionRow{
			SessionID: s.ID,
			UserID:    s.UserID,
			Mode:      s.Mode,
			Status:    s.Status,
			CreatedAt: s.CreatedAt,
			FinishedAt: s.FinishedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"test":     t,
		"sessions": out,
	})
}

