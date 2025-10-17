package handlers

import (
	"auth/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type DeleteUserHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewDeleteUserHandler(store storage.Store, logger *zap.Logger) *DeleteUserHandler {
	return &DeleteUserHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *DeleteUserHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(rw, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteUser(userID); err != nil {
		h.logger.Error("failed to delete user", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
