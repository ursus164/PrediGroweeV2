package handlers

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/clients"
	"quiz/internal/models"
	"quiz/internal/storage"
	"strings"
	"time"
)

type StartQuizHandler struct {
	storage     storage.Store
	logger      *zap.Logger
	statsClient *clients.StatsClient
}

func NewStartQuizHandler(store storage.Store, logger *zap.Logger, client *clients.StatsClient) *StartQuizHandler {
	return &StartQuizHandler{
		storage:     store,
		logger:      logger,
		statsClient: client,
	}
}

func writeJSONError(w http.ResponseWriter, status int, payload map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *StartQuizHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	h.logger.Info("Starting quiz session")
	userID := r.Context().Value("user_id").(int)

	mode, hours, err := h.storage.GetSecuritySettings()
	if err != nil {
		h.logger.Error("security settings error", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	switch mode {
	case "manual":
		ok, err := h.storage.IsUserApproved(userID)
		if err != nil {
			h.logger.Error("manual approval read error", zap.Error(err))
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}
		if !ok {
			writeJSONError(rw, http.StatusForbidden, map[string]interface{}{
				"error":   "approval_required",
				"message": "Account requires manual approval by an administrator.",
				"mode":    "manual",
			})
			return
		}
	case "cooldown":
		regAt, err := h.storage.UpsertAndGetRegisteredAt(userID)
		if err != nil {
			h.logger.Error("registered_at read error", zap.Error(err))
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}
		readyAt := regAt.Add(time.Duration(hours) * time.Hour)
		if time.Now().UTC().Before(readyAt) {
			left := time.Until(readyAt)
			waitSeconds := int(left.Seconds())
			if waitSeconds < 0 {
				waitSeconds = 0
			}
			writeJSONError(rw, http.StatusForbidden, map[string]interface{}{
				"error":         "cooldown_active",
				"message":       fmt.Sprintf("Please wait before starting the quiz."),
				"mode":          "cooldown",
				"cooldownHours": hours,
				"waitSeconds":   waitSeconds,
				"readyAt":       readyAt.UTC().Format(time.RFC3339),
			})
			return
		}
	default:
		// no restriction
	}

	var payload models.StartQuizPayload
	if err := payload.FromJSON(r.Body); err != nil {
		http.Error(rw, "invalid request payload", http.StatusBadRequest)
		return
	}
	if err := payload.Validate(); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	testCode := strings.TrimSpace(payload.TestCode)
	var newQuizSession models.QuizSession

	if testCode != "" {
		t, err := h.storage.GetTestByCode(testCode)
		if err != nil {
			h.logger.Error("failed to get test by code", zap.Error(err))
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}
		if t == nil {
			writeJSONError(rw, http.StatusBadRequest, map[string]interface{}{
				"error":   "invalid_test_code",
				"message": "Unknown test code.",
			})
			return
		}

		order, err := h.storage.GetTestQuestionIDsOrdered(t.ID)
		if err != nil {
			h.logger.Error("failed to get test questions", zap.Error(err))
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}
		if len(order) == 0 {
			h.logger.Error("test has no questions")
			http.Error(rw, "test has no questions", http.StatusServiceUnavailable)
			return
		}

		newQuizSession = models.QuizSession{
			Mode:              payload.Mode,
			UserID:            userID,
			Status:            models.QuizStatusNotStarted,
			ScreenSize:        fmt.Sprintf("%dx%d", payload.ScreenWidth, payload.ScreenHeight),
			CurrentQuestionID: order[0],
			CurrentGroup:      0,
			GroupOrder:        order,
			TestID:            &t.ID,
			TestCode:          &t.Code,
		}

	} else {
		groupID, err := h.storage.GetNextQuestionGroupID(0)
		if err != nil {
			h.logger.Error("failed to pick first group", zap.Error(err))
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}
		order, err := h.storage.GetGroupQuestionsIDsRandomOrder(groupID)
		if err != nil {
			h.logger.Error("failed to get group questions", zap.Error(err))
			http.Error(rw, "internal server error", http.StatusInternalServerError)
			return
		}
		if len(order) == 0 {
			h.logger.Error("group has no questions")
			http.Error(rw, "no questions in group", http.StatusServiceUnavailable)
			return
		}

		newQuizSession = models.QuizSession{
			Mode:              payload.Mode,
			UserID:            userID,
			Status:            models.QuizStatusNotStarted,
			ScreenSize:        fmt.Sprintf("%dx%d", payload.ScreenWidth, payload.ScreenHeight),
			CurrentQuestionID: order[0],
			CurrentGroup:      groupID,
			GroupOrder:        order,
		}
	}

	if session, err := h.storage.GetUserLastQuizSession(userID); err == nil && session != nil {
		session.FinishedAt = session.UpdatedAt
		session.Status = models.QuizStatusFinished
		_ = h.storage.UpdateQuizSession(*session)

		if testCode == "" && session.TestID == nil && session.CurrentQuestionID > 0 && len(session.GroupOrder) > 0 {
			h.logger.Info("Resuming previous non-test session ordering",
				zap.Int("prev_session_id", session.ID),
				zap.Int("prev_current_q", session.CurrentQuestionID),
				zap.Int("prev_group", session.CurrentGroup),
			)
			newQuizSession.CurrentQuestionID = session.CurrentQuestionID
			newQuizSession.CurrentGroup = session.CurrentGroup
			newQuizSession.GroupOrder = session.GroupOrder
		}
	}

	if newQuizSession.CurrentQuestionID <= 0 || len(newQuizSession.GroupOrder) == 0 {
		h.logger.Warn("new session had invalid starting state, regenerating first group/order")
		gid, err := h.storage.GetNextQuestionGroupID(0)
		if err == nil {
			if ord, e2 := h.storage.GetGroupQuestionsIDsRandomOrder(gid); e2 == nil && len(ord) > 0 {
				newQuizSession.CurrentGroup = gid
				newQuizSession.GroupOrder = ord
				newQuizSession.CurrentQuestionID = ord[0]
			}
		}
	}

	sessionCreated, err := h.storage.CreateQuizSession(newQuizSession)
	if err != nil {
		h.logger.Error("failed to create quiz session in db", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.statsClient.SaveSession(sessionCreated); err != nil {
		h.logger.Error("failed to save session in stats service", zap.Error(err))
	}

	timeLimit, err := h.storage.GetTimeLimit()
	if err != nil {
		h.logger.Error("failed to get time limit", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"session":    sessionCreated,
		"time_limit": timeLimit,
	}
	if err := json.NewEncoder(rw).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
}
