package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "go.uber.org/zap"
    "quiz/internal/storage"
)

type GetDifficultySummaryHandler struct {
    store  storage.Store
    logger *zap.Logger
}
func NewGetDifficultySummaryHandler(store storage.Store, logger *zap.Logger) *GetDifficultySummaryHandler {
    return &GetDifficultySummaryHandler{store: store, logger: logger}
}
func (h *GetDifficultySummaryHandler) Handle(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    qID, err := strconv.Atoi(idStr)
    if err != nil || qID <= 0 { http.Error(w, "bad question id", http.StatusBadRequest); return }

    s, err := h.store.GetDifficultySummary(qID)
    if err != nil { http.Error(w, "internal error", http.StatusInternalServerError); return }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(s)
}

