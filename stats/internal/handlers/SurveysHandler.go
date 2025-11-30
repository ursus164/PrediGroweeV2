package handlers

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"stats/internal/models"
	"stats/internal/storage"
	"strconv"
)

// SurveysHandler handles HTTP requests for user survey data.
type SurveysHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

// NewSurveysHandler creates a new SurveysHandler instance.
func NewSurveysHandler(storage storage.Storage, logger *zap.Logger) *SurveysHandler {
	return &SurveysHandler{
		storage: storage,
		logger:  logger,
	}
}

// Save saves a user survey response.
func (h *SurveysHandler) Save(w http.ResponseWriter, r *http.Request) {
	var surveyResponse models.SurveyResponse
	err := surveyResponse.FromJSON(r.Body)
	if err != nil {
		h.logger.Error("failed to parse survey response", zap.Error(err))
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	fmt.Println(surveyResponse)
	userID := r.Context().Value("user_id").(int)
	_, err = h.storage.GetSurveyResponseForUser(userID)
	if err == nil {
		http.Error(w, "survey response already exists", http.StatusConflict)
		return
	}
	surveyResponse.UserID = userID
	err = h.storage.SaveSurveyResponse(&surveyResponse)
	if err != nil {
		h.logger.Error("failed to save survey response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// GetSurvey retrieves survey responses for a user or all users.
func (h *SurveysHandler) GetSurvey(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		userID = strconv.Itoa(r.Context().Value("user_id").(int))
		if userID == "" {
			http.Error(w, "missing user id", http.StatusBadRequest)
			return
		}
	}

	var surveyResponses interface{}
	var err error

	if userID == "-" {
		surveyResponses, err = h.storage.GetAllSurveyResponses()
	} else {
		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}
		surveyResponses, err = h.storage.GetSurveyResponseForUser(userIDInt)
		if err != nil {
			h.logger.Error("failed to get survey response for user", zap.Error(err))
			http.Error(w, "failed to get survey response", http.StatusInternalServerError)
			return
		}
	}

	if err != nil {
		h.logger.Error("failed to get survey response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(surveyResponses)
	if err != nil {
		h.logger.Error("failed to marshal response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(res); err != nil {
		h.logger.Error("failed to write response", zap.Error(err))
	}
}
