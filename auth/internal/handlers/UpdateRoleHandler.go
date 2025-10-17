package handlers

import (
	"auth/internal/models"
	"auth/internal/storage"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type UpdateRoleHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewUpdateRoleHandler(store storage.Store, logger *zap.Logger) *UpdateRoleHandler {
	return &UpdateRoleHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *UpdateRoleHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	roleID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(rw, "invalid role id", http.StatusBadRequest)
		return
	}

	var updatedRole models.Role
	if err := json.NewDecoder(r.Body).Decode(&updatedRole); err != nil {
		http.Error(rw, "invalid request payload", http.StatusBadRequest)
		return
	}

	updatedRole.ID = roleID
	if err := h.storage.UpdateRole(updatedRole); err != nil {
		h.logger.Error("failed to update role", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
