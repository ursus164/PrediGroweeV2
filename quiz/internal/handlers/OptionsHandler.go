package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/models"
	"quiz/internal/storage"
	"strconv"
)

type OptionsHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewOptionsHandler(store storage.Store, logger *zap.Logger) *OptionsHandler {
	return &OptionsHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *OptionsHandler) GetAllOptions(w http.ResponseWriter, _ *http.Request) {
	options, err := h.storage.GetAllOptions()
	if err != nil {
		h.logger.Error("Failed to get options", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(options)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

func (h *OptionsHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	var newOption models.Option
	if err := json.NewDecoder(r.Body).Decode(&newOption); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdOption, err := h.storage.CreateOption(newOption)
	if err != nil {
		h.logger.Error("Failed to create option", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdOption)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *OptionsHandler) UpdateOption(w http.ResponseWriter, r *http.Request) {
	optionId := r.PathValue("id")
	optionID, err := strconv.Atoi(optionId)
	if err != nil {
		http.Error(w, "Invalid option ID", http.StatusBadRequest)
		return
	}
	var updatedOption models.Option
	if err := json.NewDecoder(r.Body).Decode(&updatedOption); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = h.storage.UpdateOption(optionID, updatedOption)
	if err != nil {
		h.logger.Error("Failed to update option", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *OptionsHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	optionId := r.PathValue("id")
	optionID, err := strconv.Atoi(optionId)
	if err != nil {
		http.Error(w, "Invalid option ID", http.StatusBadRequest)
		return
	}

	err = h.storage.DeleteOption(optionID)
	if err != nil {
		h.logger.Error("Failed to delete option", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
