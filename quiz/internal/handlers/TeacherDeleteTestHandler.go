package handlers

import (
    "database/sql"
    "go.uber.org/zap"
    "net/http"
    "quiz/internal/storage"
    "strconv"
)

type TeacherDeleteTestHandler struct {
    storage storage.Store
    logger  *zap.Logger
}

func NewTeacherDeleteTestHandler(store storage.Store, logger *zap.Logger) *TeacherDeleteTestHandler {
    return &TeacherDeleteTestHandler{
        storage: store,
        logger:  logger,
    }
}

func (h *TeacherDeleteTestHandler) Handle(rw http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    if idStr == "" {
        http.Error(rw, "missing id", http.StatusBadRequest)
        return
    }
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        http.Error(rw, "invalid id", http.StatusBadRequest)
        return
    }

    if err := h.storage.DeleteTest(id); err != nil {
        if err == sql.ErrNoRows {
            http.Error(rw, "not found", http.StatusNotFound)
            return
        }
        h.logger.Error("failed to delete test", zap.Error(err))
        http.Error(rw, "internal server error", http.StatusInternalServerError)
        return
    }

    rw.WriteHeader(http.StatusNoContent)
}

