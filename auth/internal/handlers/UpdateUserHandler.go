package handlers

import (
	"auth/internal/models"
	"auth/internal/storage"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// UpdateUserHandler updates user details
type UpdateUserHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewUpdateUserHandler(store storage.Store, logger *zap.Logger) *UpdateUserHandler {
	return &UpdateUserHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *UpdateUserHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(rw, "invalid user id", http.StatusBadRequest)
		return
	}
	if r.Context().Value("user_id") != userID {
		http.Error(rw, "forbidden", http.StatusForbidden)
		return
	}

	var userPayload models.UserUpdatePayload
	if err := userPayload.FromJSON(r.Body); err != nil {
		http.Error(rw, "invalid request payload", http.StatusBadRequest)
		return
	}
	dbUser, err := h.storage.GetUserById(userID, true)
	if err != nil {
		h.logger.Error("failed to get user", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
	userToUpdate := applyUpdates(dbUser, &userPayload)

	if err := h.storage.UpdateUser(userToUpdate); err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func applyUpdates(existing *models.User, updates *models.UserUpdatePayload) *models.User {
	result := existing

	if updates.FirstName != nil {
		result.FirstName = *updates.FirstName
	}
	if updates.LastName != nil {
		result.LastName = *updates.LastName
	}
	if updates.Pwd != nil {
		result.Password = *updates.Pwd
	}

	return result
}
