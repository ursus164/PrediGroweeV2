package handlers

import (
	"admin/clients"
	"admin/internal/models"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type UsersHandler struct {
	logger      *zap.Logger
	authClient  clients.AuthClient
	statsClient clients.StatsClient
}

func NewUsersHandler(logger *zap.Logger, authClient clients.AuthClient, statsClient clients.StatsClient) *UsersHandler {
	return &UsersHandler{
		logger:      logger,
		authClient:  authClient,
		statsClient: statsClient,
	}
}
func (u *UsersHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := u.authClient.GetUsers()
	if err != nil {
		u.logger.Error("failed to get users", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	usersJson, _ := json.Marshal(users)
	_, err = w.Write(usersJson)
	if err != nil {
		u.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
func (u *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user models.UserPayload
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		u.logger.Error("failed to decode request body", zap.Error(err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if user.ID == "" {
		user.ID = r.PathValue("id")
	}
	err = u.authClient.UpdateUser(user)
	if err != nil {
		u.logger.Error("failed to update user", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (u *UsersHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var userDetails models.UserDetails
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		user, err := u.authClient.GetUser(userID)
		if err != nil {
			u.logger.Error("failed to get user", zap.Error(err))
			return
		}
		userDetails.User = user
	}()
	go func() {
		defer wg.Done()
		userStats, err := u.statsClient.GetUserStats(userID)
		if err != nil {
			u.logger.Error("failed to get user stats", zap.Error(err))
			return
		}
		userDetails.Stats = userStats
	}()
	go func() {
		defer wg.Done()
		userSurvey, err := u.statsClient.GetSurvey(userID)
		if err != nil {
			u.logger.Error("failed to get user survey", zap.Error(err))
			return
		}
		userDetails.SurveyResponses = userSurvey
	}()
	wg.Wait()
	w.Header().Set("Content-Type", "application/json")
	userDetailsJson, _ := json.Marshal(userDetails)
	_, err := w.Write(userDetailsJson)
	if err != nil {
		u.logger.Error("failed to write response", zap.Error(err))
		return
	}
}

func (u *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	query := r.URL.Query().Get("withResponses")
	if query == "true" {
		err := u.statsClient.DeleteUserResponses(userID)
		if err != nil {
			u.logger.Error("failed to delete user responses", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}
	err := u.authClient.DeleteUser(userID)
	if err != nil {
		u.logger.Error("failed to delete user", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (u *UsersHandler) GetAllUsersSurveys(w http.ResponseWriter, _ *http.Request) {
	surveys, err := u.statsClient.GetAllSurveys()
	if err != nil {
		u.logger.Error("failed to get surveys", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	surveysJson, _ := json.Marshal(surveys)
	_, err = w.Write(surveysJson)
	if err != nil {
		u.logger.Error("failed to write response", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
