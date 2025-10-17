package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"quiz/internal/clients"
	"quiz/internal/middleware"
	"quiz/internal/storage"
)

type FavoriteHandler struct {
	store      storage.Store
	logger     *zap.Logger
	authClient *clients.AuthClient
}

type favNotePayload struct {
	Note *string `json:"note"`
}

func NewFavoriteHandler(store storage.Store, logger *zap.Logger, auth *clients.AuthClient) *FavoriteHandler {
	return &FavoriteHandler{store: store, logger: logger, authClient: auth}
}

func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID, err := strconv.Atoi(r.PathValue("caseId"))
	if err != nil || caseID <= 0 {
		http.Error(w, "invalid case id", http.StatusBadRequest)
		return
	}

	accessToken, err := middleware.ExtractAccessTokenFromRequest(r)
	if err != nil || accessToken == "" {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	ud, err := h.authClient.VerifyAuthToken(accessToken)
	if err != nil || ud.UserID <= 0 {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	if err := h.store.AddFavoriteCase(ud.UserID, caseID); err != nil {
		h.logger.Error("add favorite failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID, err := strconv.Atoi(r.PathValue("caseId"))
	if err != nil || caseID <= 0 {
		http.Error(w, "invalid case id", http.StatusBadRequest)
		return
	}

	accessToken, err := middleware.ExtractAccessTokenFromRequest(r)
	if err != nil || accessToken == "" {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	ud, err := h.authClient.VerifyAuthToken(accessToken)
	if err != nil || ud.UserID <= 0 {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	if err := h.store.RemoveFavoriteCase(ud.UserID, caseID); err != nil {
		h.logger.Error("remove favorite failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	accessToken, err := middleware.ExtractAccessTokenFromRequest(r)
	if err != nil || accessToken == "" {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	ud, err := h.authClient.VerifyAuthToken(accessToken)
	if err != nil || ud.UserID <= 0 {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	targetUserID := ud.UserID

	if qs := strings.TrimSpace(r.URL.Query().Get("userId")); qs != "" {
		if otherID, convErr := strconv.Atoi(qs); convErr == nil && otherID > 0 {
			role := strings.ToLower(strings.TrimSpace(fmt.Sprint(ud.Role)))
			if role != "admin" && role != "teacher" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			targetUserID = otherID
		}
	}

	data, err := h.store.ListFavoriteCases(targetUserID)
	if err != nil {
		h.logger.Error("list favorites failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(data)
}

func (h *FavoriteHandler) SetNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID, err := strconv.Atoi(r.PathValue("caseId"))
	if err != nil || caseID <= 0 {
		http.Error(w, "invalid case id", http.StatusBadRequest)
		return
	}

	accessToken, err := middleware.ExtractAccessTokenFromRequest(r)
	if err != nil || accessToken == "" {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	ud, err := h.authClient.VerifyAuthToken(accessToken)
	if err != nil || ud.UserID <= 0 {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	var p favNotePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if p.Note != nil {
		trimmed := strings.TrimSpace(*p.Note)
		if len(trimmed) == 0 {
			p.Note = nil
		} else {
			if len(trimmed) > 4000 {
				trimmed = trimmed[:4000]
			}
			p.Note = &trimmed
		}
	}

	if err := h.store.UpdateFavoriteNote(ud.UserID, caseID, p.Note); err != nil {
		h.logger.Error("update favorite note failed", zap.Error(err))
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

