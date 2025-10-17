package handlers

import (
	"auth/internal/storage"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type GetAllRolesHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewGetAllRolesHandler(store storage.Store, logger *zap.Logger) *GetAllRolesHandler {
	return &GetAllRolesHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *GetAllRolesHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	roles, err := h.storage.GetAllRoles()
	if err != nil {
		h.logger.Error("failed to get roles from db", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(roles)
}
