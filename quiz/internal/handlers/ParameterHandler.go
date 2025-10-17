package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/models"
	"quiz/internal/storage"
	"strconv"
)

// ParameterHandler handles operations on parameters
type ParameterHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewParameterHandler(store storage.Store, logger *zap.Logger) *ParameterHandler {
	return &ParameterHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *ParameterHandler) CreateParameter(w http.ResponseWriter, r *http.Request) {
	var newParameter models.Parameter
	if err := json.NewDecoder(r.Body).Decode(&newParameter); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	createdParameter, err := h.storage.CreateParameter(newParameter)
	if err != nil {
		h.logger.Error("Failed to create parameter", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdParameter)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *ParameterHandler) UpdateParameter(w http.ResponseWriter, r *http.Request) {
	parameterID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid parameter ID", http.StatusBadRequest)
		return
	}

	var updatedParameter models.Parameter
	if err := json.NewDecoder(r.Body).Decode(&updatedParameter); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedParameter.ID = parameterID
	if err := h.storage.UpdateParameter(updatedParameter); err != nil {
		h.logger.Error("Failed to update parameter", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ParameterHandler) DeleteParameter(w http.ResponseWriter, r *http.Request) {
	parameterID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid parameter ID", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteParameter(parameterID); err != nil {
		h.logger.Error("Failed to delete parameter", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ParameterHandler) GetAllParameters(w http.ResponseWriter, _ *http.Request) {
	parameters, err := h.storage.GetAllParameters()
	if err != nil {
		h.logger.Error("Failed to get parameters", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(parameters)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
func (h *ParameterHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var params []models.Parameter
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.storage.UpdateParametersOrder(params); err != nil {
		h.logger.Error("Failed to update parameters order", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
