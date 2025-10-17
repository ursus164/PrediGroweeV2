package handlers

import (
	"auth/internal/storage"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type GetAllUsersHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewGetAllUsersHandler(store storage.Store, logger *zap.Logger) *GetAllUsersHandler {
	return &GetAllUsersHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *GetAllUsersHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	users, err := h.storage.GetAllUsers()
	if err != nil {
		h.logger.Error("failed to get users from db", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(users)
}
