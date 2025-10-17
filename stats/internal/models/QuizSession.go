package models

import (
	"encoding/json"
	"io"
	"time"
)

type QuizSession struct {
	SessionID  int        `json:"session_id"`
	UserID     int        `json:"user_id"`
	FinishTime *time.Time `json:"finish_time"`
	QuizMode   string     `json:"quiz_mode"`
}

func (q *QuizSession) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(q)
}
