package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type VerifySessionHandler struct {
	logger *zap.Logger
}

func NewVerifySessionHandler(logger *zap.Logger) *VerifySessionHandler {
	return &VerifySessionHandler{
		logger: logger,
	}
}

func (h *VerifySessionHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	userRole := r.Context().Value("user_role")

	resp := map[string]interface{}{
		"user_id": userID,
		"role":    userRole,
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	err := json.NewEncoder(rw).Encode(resp)
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}
}
