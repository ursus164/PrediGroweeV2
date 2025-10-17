package models

import (
	"encoding/json"
	"io"
	"time"
)

type QuizSession struct {
	ID                    int        `json:"session_id"`
	UserID                int        `json:"user_id"`
	Mode                  QuizMode   `json:"quiz_mode"`
	Status                QuizStatus `json:"-"`
	ScreenSize            string     `json:"-"`
	CurrentQuestionID     int        `json:"-"`
	CurrentGroup          int        `json:"-"`
	GroupOrder            []int      `json:"-"`
	CreatedAt             *time.Time `json:"-"`
	UpdatedAt             *time.Time `json:"-"`
	FinishedAt            *time.Time `json:"-"`
	QuestionRequestedTime time.Time  `json:"-"`
	TestID   *int    `json:"test_id,omitempty"`
        TestCode *string `json:"test_code,omitempty"`
}

func (qs *QuizSession) ToJSON(writer io.Writer) error {
	return json.NewEncoder(writer).Encode(qs)
}
