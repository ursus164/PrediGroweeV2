package handlers

import (
	"auth/internal/auth"
	"auth/internal/models"
	"auth/internal/storage"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type OauthLoginHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewOauthLoginHandler(store storage.Store, logger *zap.Logger) *OauthLoginHandler {
	return &OauthLoginHandler{
		store:  store,
		logger: logger,
	}
}

func (h *OauthLoginHandler) HandleGoogle(w http.ResponseWriter, r *http.Request) {
	var payload models.GoogleTokenPayload
	if err := payload.FromJSON(r.Body); err != nil {
		h.logger.Error("Error decoding request body", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userInfo, err := h.verifyGoogleToken(payload.Token)
	if err != nil {
		h.logger.Error("Error verifying Google token", zap.Error(err))
		http.Error(w, "Invalid Google token", http.StatusUnauthorized)
		return
	}

	dbUser, firstLogin, err := h.findOrCreateUser(userInfo)
	if err != nil {
		h.logger.Error("Error finding/creating user", zap.Error(err))
		http.Error(w, "Error processing user", http.StatusInternalServerError)
		return
	}

	sessionId, err := auth.GenerateSessionID(64)
	if err != nil {
		h.logger.Error("Error generating session id", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = h.store.SaveUserSession(models.UserSession{
		UserID:     dbUser.ID,
		SessionID:  sessionId,
		Expiration: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		h.logger.Error("Error saving session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	accessToken, err := auth.GenerateAccessToken(strconv.Itoa(dbUser.ID))
	if err != nil {
		h.logger.Error("Error generating access token", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Path:     "/",
		Name:     "session_id",
		Value:    sessionId,
		HttpOnly: true,
		Secure:   false, // Set to true if using HTTPS
	})

	w.Header().Set("Content-Type", "application/json")
	data := map[string]interface{}{
		"user_id":      dbUser.ID,
		"role":         dbUser.Role,
		"access_token": accessToken,
		"first_login":  firstLogin,
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Error encoding response", zap.Error(err))
		return
	}
}

func (h *OauthLoginHandler) verifyGoogleToken(token string) (*models.GoogleUserInfo, error) {
	fmt.Println("google token", token)
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo?alt=json&access_token="+token, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed to get user info from google: %d", resp.StatusCode))
	}
	defer resp.Body.Close()

	var userInfo models.GoogleUserInfo
	if err = userInfo.FromJSON(resp.Body); err != nil {
		return nil, err
	}
	fmt.Println("user info from google", userInfo)

	return &userInfo, nil
}

func (h *OauthLoginHandler) findOrCreateUser(googleInfo *models.GoogleUserInfo) (*models.User, bool, error) {
	dbUser, err := h.store.GetUserByEmail(googleInfo.Email)
	if err == nil {
		if dbUser.GoogleID != googleInfo.ID {
			dbUser.GoogleID = googleInfo.ID
		}
		if dbUser.FirstName != googleInfo.FirstName {
			dbUser.FirstName = googleInfo.FirstName
		}
		if dbUser.LastName != googleInfo.LastName {
			dbUser.LastName = googleInfo.LastName
		}
		if err = h.store.UpdateUser(dbUser); err != nil {
			return nil, false, err
		}
		return dbUser, false, nil
	}
	newUser := &models.User{
		Email:     googleInfo.Email,
		GoogleID:  googleInfo.ID,
		FirstName: googleInfo.FirstName,
		LastName:  googleInfo.LastName,
		Role:      models.RoleUser,
	}
	createdUser, err := h.store.CreateUser(newUser)
	if err != nil {
		return nil, false, err
	}
	return createdUser, true, nil
}
