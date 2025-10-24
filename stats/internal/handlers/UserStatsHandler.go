package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"stats/internal/models"
	"stats/internal/storage"
	"strconv"
	"strings"
)

const (
	ctxUserIDKey        = "user_id"
	ctxRoleKeyPrimary   = "user_role" // VerifyToken wkłada tu rolę
	ctxRoleKeySecondary = "role"      // fallback
)

// UserStatsHandler handles HTTP requests for user statistics.
type UserStatsHandler struct {
	storage storage.Storage
	logger  *zap.Logger
}

// NewUserStatsHandler creates a new UserStatsHandler instance.
func NewUserStatsHandler(storage storage.Storage, logger *zap.Logger) *UserStatsHandler {
	return &UserStatsHandler{storage: storage, logger: logger}
}

// Czyta rolę z contextu niezależnie od typu (string, enum, alias itp.)
func readRole(r *http.Request) string {
	if v := r.Context().Value(ctxRoleKeyPrimary); v != nil {
		return strings.TrimSpace(fmt.Sprint(v))
	}
	if v := r.Context().Value(ctxRoleKeySecondary); v != nil {
		return strings.TrimSpace(fmt.Sprint(v))
	}
	return ""
}

func hasElevatedAccess(role string) bool {
	role = strings.TrimSpace(role)
	return strings.EqualFold(role, "admin") || strings.EqualFold(role, "teacher")
}

func (h *UserStatsHandler) resolveTargetUser(r *http.Request, role string) (int, int, error) {
	if idStr := r.PathValue("id"); idStr != "" {
		uid, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			return 0, http.StatusBadRequest, fmt.Errorf("invalid user id")
		}
		return uid, 0, nil
	}

	if q := strings.TrimSpace(r.URL.Query().Get("userId")); q != "" {
		if !hasElevatedAccess(role) {
			return 0, http.StatusForbidden, fmt.Errorf("forbidden")
		}
		uid, err := strconv.Atoi(q)
		if err != nil {
			return 0, http.StatusBadRequest, fmt.Errorf("invalid user id")
		}
		return uid, 0, nil
	}
	userID, ok := r.Context().Value(ctxUserIDKey).(int)
	if !ok {
		return 0, http.StatusUnauthorized, fmt.Errorf("missing user in context")
	}
	return userID, 0, nil
}

// Handle retrieves user statistics for the authenticated user or specified user.
func (h *UserStatsHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	role := readRole(r)

	userID, statusOverride, err := h.resolveTargetUser(r, role)
	if err != nil {
		code := statusOverride
		if code == 0 {
			code = http.StatusInternalServerError
		}
		http.Error(rw, err.Error(), code)
		return
	}

	stats := models.UserStats{
		TotalQuestions: make(map[models.QuizMode]int),
		CorrectAnswers: make(map[models.QuizMode]int),
		Accuracy:       make(map[models.QuizMode]float64),
	}

	for _, mode := range []string{models.QuizModeEducational, models.QuizModeClassic, models.QuizModeLimitedTime} {
		correct, wrong, err := h.storage.GetUserStatsForMode(userID, mode)
		if errors.Is(err, storage.ErrStatsNotFound) {
			continue
		}
		if err != nil {
			h.logger.Error("failed to get stats for user",
				zap.Int("user_id", userID),
				zap.String("mode", mode),
				zap.Error(err),
			)
			http.Error(rw, "failed to get statistics", http.StatusInternalServerError)
			return
		}
		stats.TotalQuestions[mode] = correct + wrong
		stats.CorrectAnswers[mode] = correct
		if stats.TotalQuestions[mode] != 0 {
			stats.Accuracy[mode] = float64(correct) / float64(correct+wrong)
		} else {
			stats.Accuracy[mode] = 0
		}
	}

	if err := stats.ToJSON(rw); err != nil {
		h.logger.Error("failed to encode stats", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
}

// GetUserSessions retrieves all quiz sessions for a user.
func (h *UserStatsHandler) GetUserSessions(rw http.ResponseWriter, r *http.Request) {
	role := readRole(r)

	userID, statusOverride, err := h.resolveTargetUser(r, role)
	if err != nil {
		code := statusOverride
		if code == 0 {
			code = http.StatusInternalServerError
		}
		http.Error(rw, err.Error(), code)
		return
	}

	stats, err := h.storage.GetUserQuizSessionsStats(userID)
	if err != nil {
		h.logger.Error("failed to get user sessions", zap.Error(err), zap.Int("user_id", userID))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(stats); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
}

// DeleteUserResponses deletes all responses for a specific user.
func (h *UserStatsHandler) DeleteUserResponses(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteUserResponses(userID); err != nil {
		h.logger.Error("failed to delete user responses", zap.Error(err), zap.Int("user_id", userID))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAllUsersStats retrieves statistics for all users.
func (h *UserStatsHandler) GetAllUsersStats(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.storage.GetAllUsersStats()
	if err != nil {
		h.logger.Error("failed to get user stats", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(stats)
}
