package handlers

import (
	"auth/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type DeleteRoleHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewDeleteRoleHandler(store storage.Store, logger *zap.Logger) *DeleteRoleHandler {
	return &DeleteRoleHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *DeleteRoleHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	roleID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(rw, "invalid role id", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteRole(roleID); err != nil {
		h.logger.Error("failed to delete role", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
