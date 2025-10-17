package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "go.uber.org/zap"
    "quiz/internal/storage"
)

type GetMyDifficultyHandler struct {
    store  storage.Store
    logger *zap.Logger
}
func NewGetMyDifficultyHandler(store storage.Store, logger *zap.Logger) *GetMyDifficultyHandler {
    return &GetMyDifficultyHandler{store: store, logger: logger}
}
func (h *GetMyDifficultyHandler) Handle(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    qID, err := strconv.Atoi(idStr)
    if err != nil || qID <= 0 { http.Error(w, "bad question id", http.StatusBadRequest); return }

    userID, _ := r.Context().Value("user_id").(int)
    if userID <= 0 { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

    v, err := h.store.GetMyDifficultyVote(qID, userID)
    if err != nil { http.Error(w, "internal error", http.StatusInternalServerError); return }

    w.Header().Set("Content-Type", "application/json")
    if v == nil {
        _ = json.NewEncoder(w).Encode(map[string]any{"difficulty": nil}); return
    }
    _ = json.NewEncoder(w).Encode(map[string]any{"difficulty": v.Difficulty})
}

