package api

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"net/http"
)

type QuestionImagesHandler struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewQuestionImagesHandler(logger *zap.Logger, db *sql.DB) *QuestionImagesHandler {
	return &QuestionImagesHandler{
		logger: logger,
		db:     db,
	}
}

func (h *QuestionImagesHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	questionId := r.PathValue("questionId")
	h.logger.Info("Getting image for question id: " + questionId)
	questionID, err := strconv.Atoi(questionId)
	if err != nil {
		http.Error(rw, "Invalid question id", http.StatusBadRequest)
		return
	}
	id := r.PathValue("id")
	if id != "1" && id != "2" && id != "3" {
		http.Error(rw, "Invalid image id", http.StatusBadRequest)
		return
	}
	var imagePath string
	err = h.db.QueryRow("SELECT image"+id+"_path FROM question_images WHERE question_id = $1", questionID).Scan(&imagePath)
	if err != nil {
		http.Error(rw, "Failed to get image", http.StatusInternalServerError)
		return
	}
	imagePath = strings.Replace(imagePath, "\\", "/", -1)
	imagePath = filepath.Clean(imagePath)

	fullPath := filepath.Join("/app/images", imagePath)

	fmt.Println("serving from: ", fullPath)
	ext := strings.ToLower(filepath.Ext(fullPath))
	contentType := "image/jpeg"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}
	rw.Header().Set("Content-Type", contentType)

	//debug
	if info, err := os.Stat(fullPath); err != nil {
		fmt.Printf("File stat error: %v\n", err)
	} else {
		fmt.Printf("File permissions: %v\n", info.Mode())
	}

	http.ServeFile(rw, r, fullPath)

}
