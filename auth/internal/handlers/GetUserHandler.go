package handlers

import (
	"auth/internal/storage"
	"go.uber.org/zap"
	"net/http"
)

type GetUserHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewGetUserHandler(store storage.Store, logger *zap.Logger) *GetUserHandler {
	return &GetUserHandler{
		logger: logger,
		store:  store,
	}
}
func (h *GetUserHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	user, err := h.store.GetUserById(userID, false)
	if err != nil {
		http.Error(rw, "User not found", http.StatusNotFound)
		return
	}
	rw.WriteHeader(http.StatusOK)
	err = user.ToJSON(rw)
	if err != nil {
		h.logger.Error("Error marshalling user", zap.Error(err))
	}
}
