package models

import (
	"encoding/json"
	"io"
	"time"
)

// QuestionResponse represents a user's answer to a quiz question.
type QuestionResponse struct {
	ID         int        `json:"id"`
	QuestionID int        `json:"question_id"`
	CaseCode   string     `json:"case_code"`
	Answer     string     `json:"answer"`
	IsCorrect  bool       `json:"is_correct"`
	Time       *time.Time `json:"time,omitempty"`
	UserID     *int       `json:"user_id,omitempty"`
	ScreenSize string     `json:"screen_size"`
	TimeSpent  int        `json:"time_spent"`
}

// FromJSON decodes a QuestionResponse from JSON.
func (q *QuestionResponse) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(q)
}
