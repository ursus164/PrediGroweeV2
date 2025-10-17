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

type SurveysHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

func NewSurveysHandler(storage storage.Storage, logger *zap.Logger) *SurveysHandler {
	return &SurveysHandler{
		storage: storage,
		logger:  logger,
	}
}

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

func (h *SurveysHandler) GetSurvey(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	if userId == "" {
		userId = strconv.Itoa(r.Context().Value("user_id").(int))
		if userId == "" {
			http.Error(w, "missing user id", http.StatusBadRequest)
			return
		}
	}

	var surveyResponses interface{}
	var err error

	if userId == "-" {
		surveyResponses, err = h.storage.GetAllSurveyResponses()
	} else {
		userID, err := strconv.Atoi(userId)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}
		surveyResponses, err = h.storage.GetSurveyResponseForUser(userID)
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
