package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type VerifyTokenHandler struct {
}

func NewVerifyTokenHandler() *VerifyTokenHandler {
	return &VerifyTokenHandler{}
}
func (h *VerifyTokenHandler) Handle(rw http.ResponseWriter, r *http.Request) {
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

	log.Println("VerifyTokenHandler completed successfully")
}
