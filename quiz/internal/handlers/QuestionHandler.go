package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/models"
	"quiz/internal/storage"
	"strconv"
)

type QuestionHandler struct {
	storage storage.Store
	logger  *zap.Logger
}

func NewQuestionHandler(store storage.Store, logger *zap.Logger) *QuestionHandler {
	return &QuestionHandler{
		storage: store,
		logger:  logger,
	}
}

// CreateQuestion creates a new question
func (h *QuestionHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var questionPayload models.QuestionPayload
	if err := questionPayload.FromJSON(r.Body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if questionPayload.CaseID == 0 {
		http.Error(w, "Invalid case ID", http.StatusBadRequest)
		return
	}
	createdQuestion, err := h.storage.CreateQuestion(questionPayload)
	if err != nil {
		h.logger.Error("Failed to create question", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = createdQuestion.ToJSON(w)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}
func (h *QuestionHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}
	var questionPayload models.Question
	if err := questionPayload.FromJSON(r.Body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	questionToUpdate := models.QuestionPayload{
		Question:      questionPayload.Question,
		Answers:       questionPayload.Options,
		CaseID:        questionPayload.Case.ID,
		PredictionAge: questionPayload.PredictionAge,
		Group:         questionPayload.Group,
	}
	_, err = h.storage.UpdateQuestionByID(questionID, questionToUpdate)
	if err != nil {
		h.logger.Error("Failed to update question", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if questionPayload.Correct != nil {
		err = h.storage.UpdateQuestionCorrectOption(questionID, *questionPayload.Correct)
		if err != nil {
			h.logger.Error("Failed to update question correct option", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	casePayload := questionPayload.Case
	if err := h.storage.UpdateCaseParameters(casePayload.ID, casePayload.Parameters, casePayload.ParameterValues); err != nil {
		http.Error(w, "Failed to update case parameters", http.StatusInternalServerError)
		return
	}

}
func (h *QuestionHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}
	err = h.storage.DeleteQuestionByID(questionID)
	if err != nil {
		h.logger.Error("Failed to delete question", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *QuestionHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}
	question, err := h.storage.GetQuestionByID(questionID)
	if err != nil {
		h.logger.Error("Failed to get question", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	correct, err := h.storage.GetQuestionCorrectOption(questionID)
	if err != nil {
		h.logger.Error("Failed to get question correct option", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	question.Correct = &correct
	err = question.ToJSON(w)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
func (h *QuestionHandler) GetAllQuestions(w http.ResponseWriter, _ *http.Request) {
	questions, err := h.storage.GetAllQuestions()
	if err != nil {
		h.logger.Error("Failed to get questions", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(questions)
	if err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
