package handlers

import (
	"auth/internal/models"
	"auth/internal/storage"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type CreateRoleHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewCreateRoleHandler(store storage.Store, logger *zap.Logger) *CreateRoleHandler {
	return &CreateRoleHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *CreateRoleHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	var newRole models.Role
	if err := json.NewDecoder(r.Body).Decode(&newRole); err != nil {
		http.Error(rw, "invalid request payload", http.StatusBadRequest)
		return
	}

	createdRole, err := h.storage.CreateRole(newRole)
	if err != nil {
		h.logger.Error("failed to create role", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	json.NewEncoder(rw).Encode(createdRole)
}
