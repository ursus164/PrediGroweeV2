package models

import (
	"encoding/json"
	"io"
	"time"
)

// QuizSession represents a user's quiz session.
type QuizSession struct {
	SessionID  int        `json:"session_id"`
	UserID     int        `json:"user_id"`
	FinishTime *time.Time `json:"finish_time"`
	QuizMode   string     `json:"quiz_mode"`
}

// FromJSON decodes a QuizSession from JSON.
func (q *QuizSession) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(q)
}
