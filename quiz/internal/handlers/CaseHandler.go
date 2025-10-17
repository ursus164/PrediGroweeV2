package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/models"
	"quiz/internal/storage"
	"strconv"
)

type CaseHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

type caseV3Response struct {
    Age3   int                      `json:"age3"`
    Values []models.ParameterValue  `json:"values"`
}

func NewCaseHandler(store storage.Store, logger *zap.Logger) *CaseHandler {
	return &CaseHandler{
		storage: store,
		logger:  logger,
	}
}

func (h *CaseHandler) CreateCase(w http.ResponseWriter, r *http.Request) {
	var CasePayload models.Case
	err := CasePayload.FromJSON(r.Body)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	createdCase, err := h.storage.CreateCase(CasePayload)
	if err != nil {
		h.logger.Error("Failed to create case", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	createdParametersValues := make([]models.ParameterValue, 0)
	parameters := make([]models.Parameter, 0)
	for _, pValue := range CasePayload.ParameterValues {
		createdParameter, err := h.storage.CreateCaseParameter(createdCase.ID, pValue)
		if err != nil {
			h.logger.Error("Failed to create case parameter", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		createdParametersValues = append(createdParametersValues, createdParameter)
		parameter, err := h.storage.GetParameterByID(pValue.ParameterID)
		parameters = append(parameters, parameter)
	}
	createdCase.ParameterValues = createdParametersValues
	createdCase.Parameters = parameters

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdCase)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *CaseHandler) UpdateCase(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *CaseHandler) DeleteCase(w http.ResponseWriter, r *http.Request) {
	caseID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}
	if err := h.storage.DeleteCaseWithParameters(caseID); err != nil {
		h.logger.Error("Failed to delete case", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CaseHandler) GetAllCases(w http.ResponseWriter, _ *http.Request) {
	cases, err := h.storage.GetAllCases()
	if err != nil {
		h.logger.Error("Failed to get cases", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cases)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *CaseHandler) GetCaseByID(w http.ResponseWriter, r *http.Request) {
	caseID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}
	dbCase, err := h.storage.GetCaseByID(caseID)
	if err != nil {
		h.logger.Error("Failed to get case", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = dbCase.ToJSON(w)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *CaseHandler) GetCaseParametersV3(w http.ResponseWriter, r *http.Request) {
    caseID, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, "Invalid case ID", http.StatusBadRequest)
        return
    }

    vals, err := h.storage.GetCaseParametersV3(caseID)
    if err != nil {
        h.logger.Error("Failed to get v3 parameters", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    age3, err := h.storage.GetCaseAge3(caseID)
    if err != nil {
        h.logger.Error("Failed to get age3", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(caseV3Response{
        Age3:   age3,
        Values: vals,
    })
}
