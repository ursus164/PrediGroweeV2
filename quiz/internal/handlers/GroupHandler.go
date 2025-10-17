package handlers

import (
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/storage"
)

// GroupHandler handles operations on case groups
type GroupHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewGroupHandler(store storage.Store, logger *zap.Logger) *GroupHandler {
	return &GroupHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *GroupHandler) GetAllGroups(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
