package handlers

import (
	"admin/clients"
	"admin/internal/models"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
    "strconv"
)

type QuizHandler struct {
	logger      *zap.Logger
	quizClient  clients.QuizClient
	statsClient clients.StatsClient
}

func NewQuizHandler(logger *zap.Logger, quizClient clients.QuizClient, statsClient clients.StatsClient) *QuizHandler {
	return &QuizHandler{
		logger:      logger,
		quizClient:  quizClient,
		statsClient: statsClient,
	}
}

func (h *QuizHandler) GetAllQuestions(w http.ResponseWriter, _ *http.Request) {
	questions, err := h.quizClient.GetAllQuestions()
	if err != nil {
		h.logger.Error("Failed to get questions", zap.Error(err))
		http.Error(w, "Failed to get questions", http.StatusInternalServerError)
		return
	}
	questionsJSON, err := json.Marshal(questions)
	if err != nil {
		h.logger.Error("Failed to marshal questions", zap.Error(err))
		http.Error(w, "Failed to marshal questions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(questionsJSON)
}

func (h *QuizHandler) GetAllParameters(w http.ResponseWriter, _ *http.Request) {
	parameters, err := h.quizClient.GetAllParameters()
	if err != nil {
		h.logger.Error("Failed to get parameters", zap.Error(err))
		http.Error(w, "Failed to get parameters", http.StatusInternalServerError)
		return
	}
	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		h.logger.Error("Failed to marshal parameters", zap.Error(err))
		http.Error(w, "Failed to marshal parameters", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(parametersJSON)
}

func (h *QuizHandler) UpdateParameter(w http.ResponseWriter, r *http.Request) {
	paramId := r.PathValue("id")
	var updatedParameter models.Parameter
	if err := json.NewDecoder(r.Body).Decode(&updatedParameter); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	if err := h.quizClient.UpdateParameter(paramId, updatedParameter); err != nil {
		h.logger.Error("Failed to update parameter", zap.Error(err))
		http.Error(w, "Failed to update parameter", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *QuizHandler) DeleteParameter(w http.ResponseWriter, r *http.Request) {
	paramId := r.PathValue("id")
	if paramId == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}
	if err := h.quizClient.DeleteParameter(paramId); err != nil {
		h.logger.Error("Failed to delete parameter", zap.Error(err))
		http.Error(w, "Failed to delete parameter", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *QuizHandler) GetAllOptions(w http.ResponseWriter, _ *http.Request) {
	options, err := h.quizClient.GetAllOptions()
	if err != nil {
		h.logger.Error("Failed to get options", zap.Error(err))
		http.Error(w, "Failed to get options", http.StatusInternalServerError)
		return
	}
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		h.logger.Error("Failed to marshal options", zap.Error(err))
		http.Error(w, "Failed to marshal options", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(optionsJSON)
}

func (h *QuizHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	questionId := r.PathValue("id")
	question, err := h.quizClient.GetQuestion(questionId)
	if err != nil {
		h.logger.Error("Failed to get question", zap.Error(err))
		http.Error(w, "Failed to get question", http.StatusInternalServerError)
		return
	}
	questionJSON, err := json.Marshal(question)
	if err != nil {
		h.logger.Error("Failed to marshal question", zap.Error(err))
		http.Error(w, "Failed to marshal question", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(questionJSON)
}

func (h *QuizHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionId := r.PathValue("id")
	var updatedQuestion models.Question
	if err := json.NewDecoder(r.Body).Decode(&updatedQuestion); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	if err := h.quizClient.UpdateQuestion(questionId, updatedQuestion); err != nil {
		h.logger.Error("Failed to update question", zap.Error(err))
		http.Error(w, "Failed to update question", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *QuizHandler) CreateParameter(w http.ResponseWriter, r *http.Request) {
	var newParameter models.Parameter
	if err := json.NewDecoder(r.Body).Decode(&newParameter); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	param, err := h.quizClient.CreateParameter(newParameter)
	if err != nil {
		h.logger.Error("Failed to create parameter", zap.Error(err))
		http.Error(w, "Failed to create parameter", http.StatusInternalServerError)
		return
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		h.logger.Error("Failed to marshal parameter", zap.Error(err))
		http.Error(w, "Failed to marshal parameter", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(paramJSON)
}

func (h *QuizHandler) UpdateOption(w http.ResponseWriter, r *http.Request) {
	optionId := r.PathValue("id")
	var updatedOption models.Option
	if err := json.NewDecoder(r.Body).Decode(&updatedOption); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	if err := h.quizClient.UpdateOption(optionId, updatedOption); err != nil {
		h.logger.Error("Failed to update option", zap.Error(err))
		http.Error(w, "Failed to update option", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *QuizHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	var newOption models.Option
	if err := json.NewDecoder(r.Body).Decode(&newOption); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	option, err := h.quizClient.CreateOption(newOption)
	if err != nil {
		h.logger.Error("Failed to create option", zap.Error(err))
		http.Error(w, "Failed to create option", http.StatusInternalServerError)
		return
	}
	optionJSON, err := json.Marshal(option)
	if err != nil {
		h.logger.Error("Failed to marshal option", zap.Error(err))
		http.Error(w, "Failed to marshal option", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(optionJSON)
}

func (h *QuizHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	optionId := r.PathValue("id")
	if err := h.quizClient.DeleteOption(optionId); err != nil {
		h.logger.Error("Failed to delete option", zap.Error(err))
		http.Error(w, "Failed to delete option", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *QuizHandler) UpdateParametersOrder(w http.ResponseWriter, r *http.Request) {
	var newOrder []models.Parameter
	if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	if err := h.quizClient.UpdateParametersOrder(newOrder); err != nil {
		h.logger.Error("Failed to update parameters order", zap.Error(err))
		http.Error(w, "Failed to update parameters order", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *QuizHandler) GetSettings(w http.ResponseWriter, _ *http.Request) {
	settings, err := h.quizClient.GetSettings()
	if err != nil {
		h.logger.Error("Failed to get settings", zap.Error(err))
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		h.logger.Error("Failed to marshal settings", zap.Error(err))
		http.Error(w, "Failed to marshal settings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(settingsJSON)
}

func (h *QuizHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var newSettings []models.Settings
	if err := json.NewDecoder(r.Body).Decode(&newSettings); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}
	if err := h.quizClient.UpdateSettings(newSettings); err != nil {
		h.logger.Error("Failed to update settings", zap.Error(err))
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *QuizHandler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if err := h.quizClient.ApproveUser(body.UserID); err != nil {
		h.logger.Error("Failed to approve user", zap.Error(err))
		http.Error(w, "Failed to approve user", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *QuizHandler) UnapproveUser(w http.ResponseWriter, r *http.Request) {
  var payload struct{ UserID int `json:"user_id"` }
  if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.UserID == 0 {
    http.Error(w, "invalid request body", http.StatusBadRequest); return
  }
  if err := h.quizClient.UnapproveUser(payload.UserID); err != nil {
    h.logger.Error("failed to unapprove user", zap.Error(err))
    http.Error(w, "internal server error", http.StatusInternalServerError); return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  _, _ = w.Write([]byte(`{"status":"ok"}`))
}


func (h *QuizHandler) ListActiveSessions(w http.ResponseWriter, r *http.Request) {
	cutoff := 5
	if v := r.URL.Query().Get("cutoff"); v != "" {
		if n, err := strconv.Atoi(v); err == nil { cutoff = n }
	}
	sessions, err := h.quizClient.ListActiveSessions(cutoff)
	if err != nil {
		h.logger.Error("Failed to list active sessions", zap.Error(err))
		http.Error(w, "Failed to list active sessions", http.StatusBadGateway)
		return
	}

	ids := make([]int, 0, len(sessions))
	for _, s := range sessions { ids = append(ids, s.ID) }

	accs, err := h.statsClient.GetSessionsAccuracy(ids)
	if err != nil {
		h.logger.Error("GetSessionsAccuracy failed", zap.Error(err))
	} else {
		m := make(map[int]models.SessionAccuracy, len(accs))
		for _, a := range accs { m[a.SessionID] = a }
		for i := range sessions {
			if a, ok := m[sessions[i].ID]; ok {
				acc := a.Accuracy
				cor := a.Correct
				tot := a.Total
				sessions[i].Accuracy = &acc
				sessions[i].CorrectAnswers = &cor
				sessions[i].TotalAnswers = &tot
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sessions)
}
