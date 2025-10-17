package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"go.uber.org/zap"
	"quiz/internal/storage"
)

type ApprovalHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewApprovalHandler(s storage.Store, l *zap.Logger) *ApprovalHandler {
	return &ApprovalHandler{storage: s, logger: l}
}

func (h *ApprovalHandler) Approve(w http.ResponseWriter, r *http.Request) {
	var body struct{ UserID int `json:"user_id"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if err := h.storage.ApproveUser(body.UserID, nil); err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	go func(userID int) {
		payload := map[string]any{"user_id": userID}
		b, _ := json.Marshal(payload)

		authBase := os.Getenv("AUTH_BASE_URL")
		if authBase == "" {
			authBase = "http://auth:8080"
		}
		req, err := http.NewRequest("POST", authBase+"/auth/notify-approved", bytes.NewReader(b))
		if err != nil {
			h.logger.Warn("notify-approved: build req failed", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Api-Key", os.Getenv("INTERNAL_API_KEY"))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			h.logger.Warn("notify-approved: call failed", zap.Error(err))
			return
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			h.logger.Warn("notify-approved: non-204 status", zap.Int("status", resp.StatusCode))
		}
	}(body.UserID)
	// -------------------------------------------------------

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *ApprovalHandler) Unapprove(w http.ResponseWriter, r *http.Request) {
  var body struct{ UserID int `json:"user_id"` }
  if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == 0 {
    http.Error(w, "Bad request", http.StatusBadRequest); return
  }
  if err := h.storage.UnapproveUser(body.UserID, nil); err != nil {
    http.Error(w, "DB error", http.StatusInternalServerError); return
  }
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"status":"ok"}`))
}

func (h *ApprovalHandler) GetApproved(w http.ResponseWriter, _ *http.Request) {
	ids, err := h.storage.GetApprovedUserIDs()
	if err != nil {
		h.logger.Error("failed to get approved user ids", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"approved_user_ids": ids})
}


