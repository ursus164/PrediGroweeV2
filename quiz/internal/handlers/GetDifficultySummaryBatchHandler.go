package handlers

import (
  "encoding/json"
  "net/http"
  "strconv"
  "strings"

  "go.uber.org/zap"
  "quiz/internal/storage"
)

type GetDifficultySummaryBatchHandler struct {
  store  storage.Store
  logger *zap.Logger
}

func NewGetDifficultySummaryBatchHandler(store storage.Store, logger *zap.Logger) *GetDifficultySummaryBatchHandler {
  return &GetDifficultySummaryBatchHandler{store: store, logger: logger}
}

func (h *GetDifficultySummaryBatchHandler) Handle(w http.ResponseWriter, r *http.Request) {
  idsParam := r.URL.Query().Get("ids")
  if strings.TrimSpace(idsParam) == "" {
    http.Error(w, "ids required", http.StatusBadRequest)
    return
  }
  parts := strings.Split(idsParam, ",")
  ids := make([]int, 0, len(parts))
  seen := make(map[int]struct{})
  for _, p := range parts {
    n, err := strconv.Atoi(strings.TrimSpace(p))
    if err == nil && n > 0 {
      if _, ok := seen[n]; !ok {
        seen[n] = struct{}{}
        ids = append(ids, n)
      }
    }
  }
  if len(ids) == 0 {
    _ = json.NewEncoder(w).Encode([]any{})
    return
  }

  rows, err := h.store.GetDifficultySummaryBatch(ids)
  if err != nil {
    http.Error(w, "internal error", http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  _ = json.NewEncoder(w).Encode(rows)
}

