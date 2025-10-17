package handlers

import (
    "encoding/json"
    "go.uber.org/zap"
    "net/http"
    "quiz/internal/storage"
    "strconv"
)

type TeacherGetTestHandler struct {
    storage storage.Store
    logger  *zap.Logger
}

func NewTeacherGetTestHandler(store storage.Store, logger *zap.Logger) *TeacherGetTestHandler {
    return &TeacherGetTestHandler{
        storage: store,
        logger:  logger,
    }
}

func (h *TeacherGetTestHandler) Handle(rw http.ResponseWriter, r *http.Request) {
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

    t, err := h.storage.GetTestByID(id)
    if err != nil {
        h.logger.Error("failed to get test by id", zap.Error(err))
        http.Error(rw, "not found", http.StatusNotFound)
        return
    }
    if t == nil {
        http.Error(rw, "not found", http.StatusNotFound)
        return
    }

    // Pobierz pytania testu w zadanej kolejności
    qIDs, err := h.storage.GetTestQuestionIDsOrdered(t.ID)
    if err != nil {
        h.logger.Error("failed to get test question ids", zap.Error(err))
        http.Error(rw, "internal server error", http.StatusInternalServerError)
        return
    }

    // Zbuduj lekką listę pytań (id + kilka pól), aby dialog mógł je wyświetlić
    type qLight struct {
        ID     int    `json:"id"`
        Code   string `json:"case_code"`
        Gender string `json:"gender"`
        Group  int    `json:"group"`
    }
    questions := make([]qLight, 0, len(qIDs))
    for _, qid := range qIDs {
        q, err := h.storage.GetQuestionByID(qid)
        if err != nil {
            h.logger.Warn("failed to get question for test", zap.Int("question_id", qid), zap.Error(err))
            continue
        }
        questions = append(questions, qLight{
            ID:     q.ID,
            Code:   q.Case.Code,
            Gender: q.Case.Gender,
            Group:  q.Group,
        })
    }

    resp := map[string]interface{}{
        "id":              t.ID,
        "code":            t.Code,
        "name":            t.Name,
        "created_by":      t.CreatedBy,
        "created_at":      t.CreatedAt,
        "question_ids":    qIDs,
        "questions":       questions,
        "questions_count": len(qIDs),
    }

    rw.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(rw).Encode(resp); err != nil {
        h.logger.Error("failed to encode response", zap.Error(err))
        http.Error(rw, "internal server error", http.StatusInternalServerError)
        return
    }
}

