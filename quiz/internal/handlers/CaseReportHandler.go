package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"quiz/internal/clients"
	"quiz/internal/middleware"
	"quiz/internal/storage"
)

type CaseReportHandler struct {
	store      storage.Store
	logger     *zap.Logger
	authClient *clients.AuthClient
}

func NewCaseReportHandler(store storage.Store, logger *zap.Logger, authClient *clients.AuthClient) *CaseReportHandler {
	return &CaseReportHandler{store: store, logger: logger, authClient: authClient}
}

type reportPayload struct {
	Description string `json:"description"`
}

type notePayload struct {
	Note *string `json:"note"`
}
const maxReportLen = 4000

func userIDFromCtx(ctx context.Context) (int, bool) {
	if v := ctx.Value("user_id"); v != nil {
		if id, ok := v.(int); ok && id > 0 {
			return id, true
		}
	}
	return 0, false
}

// POST /quiz/cases/{caseId}/report  (VerifyToken)
func (h *CaseReportHandler) Report(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseIDStr := r.PathValue("caseId")
	caseID, err := strconv.Atoi(caseIDStr)
	if err != nil || caseID <= 0 {
		http.Error(w, "invalid case id", http.StatusBadRequest)
		return
	}

	userID, ok := userIDFromCtx(r.Context())

	if !ok {
		accessToken, err := middleware.ExtractAccessTokenFromRequest(r)
		if err != nil || accessToken == "" {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		ud, err := h.authClient.VerifyAuthToken(accessToken)
		if err != nil || ud.UserID <= 0 {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		userID = ud.UserID
	}

	var p reportPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	desc := strings.TrimSpace(p.Description)
	if desc == "" {
		http.Error(w, "description required", http.StatusBadRequest)
		return
	}
	if len(desc) > maxReportLen {
		desc = desc[:maxReportLen]
	}

	if err := h.store.CreateCaseReport(caseID, userID, desc); err != nil {
		h.logger.Error("create case report failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

// GET /quiz/reports  (InternalAuth)
func (h *CaseReportHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reps, err := h.store.ListCaseReports()
	if err != nil {
		h.logger.Error("list case reports failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(reps)
}

// DELETE /quiz/reports/{id}  (InternalAuth)
func (h *CaseReportHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteCaseReport(id); err != nil {
		h.logger.Error("delete case report failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CaseReportHandler) SetNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var p notePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if p.Note != nil {
		trimmed := strings.TrimSpace(*p.Note)
		if len(trimmed) > maxReportLen {
			trimmed = trimmed[:maxReportLen]
		}
		p.Note = &trimmed
	}

	var adminIDPtr *int
	if aid, ok := userIDFromCtx(r.Context()); ok && aid > 0 {
		adminIDPtr = &aid
	}

	if err := h.store.UpdateCaseReportNote(id, p.Note, adminIDPtr); err != nil {
		h.logger.Error("update report note failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /quiz/reports/pendingCount  (InternalAuth)
func (h *CaseReportHandler) PendingCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	n, err := h.store.CountReportsWithoutNote()
	if err != nil {
		h.logger.Error("count pending reports failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]int{"count": n})
}


