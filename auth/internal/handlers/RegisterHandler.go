package handlers

import (
	"auth/internal/auth"
	"auth/internal/models"
	"auth/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
)

type RegisterHandler struct {
	store  storage.Store
	logger *zap.Logger
}

func NewRegisterHandler(store storage.Store, logger *zap.Logger) *RegisterHandler {
	return &RegisterHandler{
		logger: logger,
		store:  store,
	}
}

func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Registering user")
	var user models.User
	err := user.FromJSON(r.Body)
	if err != nil {
		h.logger.Error("Error unmarshalling json", zap.Error(err))
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	err = user.Validate()
	if err != nil {
		h.logger.Error("Error validating user", zap.Error(err))
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	user.Email = strings.ToLower(user.Email)
	if _, err := h.store.GetUserByEmail(user.Email); err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Error hashing password", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)
	userCreated, err := h.store.CreateUser(&user)
	if err != nil {
		h.logger.Error("Error creating user", zap.Error(err))
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	verificationToken, err := auth.GenerateVerificationToken(strconv.Itoa(userCreated.ID))
	err = auth.SendVerificationEmail(userCreated.Email, verificationToken)
	if err != nil {
		h.logger.Error("Error sending verification email", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = userCreated.ToJSON(w)
	if err != nil {
		h.logger.Error("Error marshalling user", zap.Error(err))
	}

	//accessToken, err := auth.GenerateAccessToken(strconv.Itoa(userCreated.ID))
	//if err != nil {
	//	h.logger.Error("Error generating access token", zap.Error(err))
	//	http.Error(w, "Internal server error", http.StatusInternalServerError)
	//	return
	//}
	//sessionId, err := auth.GenerateSessionID(64)
	//if err != nil {
	//	h.logger.Error("Error generating session id", zap.Error(err))
	//	http.Error(w, "Internal server error", http.StatusInternalServerError)
	//	return
	//}
	//err = h.store.SaveUserSession(models.UserSession{
	//	UserID:     userCreated.ID,
	//	SessionID:  sessionId,
	//	Expiration: time.Now().Add(7 * 24 * time.Hour),
	//})
	//if err != nil {
	//	h.logger.Error("Error saving session", zap.Error(err))
	//	http.Error(w, "Internal server error", http.StatusInternalServerError)
	//	return
	//}
	//auth.SetCookie(w, "session_id", sessionId)
	//w.Header().Set("Content-Type", "application/json")
	//w.WriteHeader(http.StatusCreated)
	//response := map[string]interface{}{
	//	"user":         userCreated,
	//	"access_token": accessToken,
	//}
	//if err := json.NewEncoder(w).Encode(response); err != nil {
	//	h.logger.Error("Error writing response", zap.Error(err))
	//	http.Error(w, "Internal server error", http.StatusInternalServerError)
	//	return
	//}
}

func (h *RegisterHandler) Verify(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Verifying user")
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	jwtToken, err := auth.ValidateJWT(token)
	if err != nil {
		h.logger.Error("Error validating token", zap.Error(err))
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	userId := jwtToken.Claims.(jwt.MapClaims)["sub"].(string)
	userID, err := strconv.Atoi(userId)
	if err != nil {
		h.logger.Error("Error converting user id", zap.Error(err))
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	user, err := h.store.GetUserByIdInternal(userID)
	if err != nil {
		h.logger.Error("Error getting user by id", zap.Error(err))
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	user.Verified = true
	err = h.store.UpdateUser(user)
	if err != nil {
		h.logger.Error("Error updating user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte("User verified"))
	if err != nil {
		h.logger.Error("Error writing response", zap.Error(err))
	}
}
