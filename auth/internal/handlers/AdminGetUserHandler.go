package handlers

import (
	"auth/internal/storage"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type AdminGetUserHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewAdminGetUserHandler(storage storage.Store, logger *zap.Logger) *AdminGetUserHandler {
	return &AdminGetUserHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *AdminGetUserHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}
	user, err := h.storage.GetUserById(userId, false)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		h.logger.Error("Error marshalling user", zap.Error(err))
	}
}
