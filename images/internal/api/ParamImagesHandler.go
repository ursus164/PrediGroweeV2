// Package api contains HTTP handlers for the Images service.
package api

import (
	"database/sql"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// ParamImagesHandler handles CRUD operations for parameter images.
type ParamImagesHandler struct {
	logger *zap.Logger
	db     *sql.DB
}

// NewParamImagesHandler creates a new ParamImagesHandler.
func NewParamImagesHandler(logger *zap.Logger, db *sql.DB) *ParamImagesHandler {
	return &ParamImagesHandler{
		logger: logger,
		db:     db,
	}
}

// GetImage returns image bytes for a given parameter id.
func (h *ParamImagesHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	var img []byte
	err := h.db.QueryRowContext(r.Context(), "SELECT image FROM params_images WHERE param_id = $1", id).Scan(&img)
	if err == sql.ErrNoRows {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to get image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png") // Adjust if you support different types
	if _, err := w.Write(img); err != nil {
		h.logger.Warn("failed to write image response", zap.Error(err))
	}
}

// PostImage upserts an image for a given parameter id.
func (h *ParamImagesHandler) PostImage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	// Max 5MB
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image from form", http.StatusBadRequest)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			h.logger.Warn("failed to close uploaded file", zap.Error(cerr))
		}
	}()

	img, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	if len(img) == 0 {
		http.Error(w, "Empty image not allowed", http.StatusBadRequest)
		return
	}

	// Upsert - update if exists, insert if not
	_, err = h.db.ExecContext(r.Context(), `
        INSERT INTO params_images (param_id, image)
        VALUES ($1, $2)
        ON CONFLICT (param_id)
        DO UPDATE SET image = EXCLUDED.image
    `, id, img)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
