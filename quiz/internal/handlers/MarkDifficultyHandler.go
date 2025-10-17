package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/lib/pq"
    "go.uber.org/zap"
    "quiz/internal/models"
    "quiz/internal/storage"
)

type MarkDifficultyHandler struct {
    store  storage.Store
    logger *zap.Logger
}
func NewMarkDifficultyHandler(store storage.Store, logger *zap.Logger) *MarkDifficultyHandler {
    return &MarkDifficultyHandler{store: store, logger: logger}
}
func (h *MarkDifficultyHandler) Handle(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    qID, err := strconv.Atoi(idStr)
    if err != nil || qID <= 0 { http.Error(w, "bad question id", http.StatusBadRequest); return }

    userID, _ := r.Context().Value("user_id").(int)
    if userID <= 0 { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

    var req struct{ Difficulty string `json:"difficulty"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest); return
    }
    if req.Difficulty != "easy" && req.Difficulty != "hard" {
        http.Error(w, "difficulty must be easy|hard", http.StatusBadRequest); return
    }

    if err := h.store.InsertDifficultyVote(qID, userID, models.DifficultyLevel(req.Difficulty)); err != nil {
        if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
            http.Error(w, "already voted", http.StatusConflict); return
        }
        h.logger.Error("insert diff vote failed", zap.Error(err))
        http.Error(w, "internal error", http.StatusInternalServerError); return
    }
    w.WriteHeader(http.StatusCreated)
}

