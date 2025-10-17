package handlers

import (
	"auth/internal/models"
	"auth/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type AdminUpdateUserHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewAdminUpdateUserHandler(store storage.Store, logger *zap.Logger) *AdminUpdateUserHandler {
	return &AdminUpdateUserHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *AdminUpdateUserHandler) Handle(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update user handler")
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	userPayload := models.UserUpdatePayload{}
	if err := userPayload.FromJSON(r.Body); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	dbUser, err := h.storage.GetUserById(userID, true)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	userToUpdate := dbUser
	userToUpdate.Role = *userPayload.Role

	if err := h.storage.UpdateUser(userToUpdate); err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
