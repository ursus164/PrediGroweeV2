package models

import "time"

type ActiveSession struct {
	ID                 int        `json:"id"`
	UserID             int        `json:"user_id"`
	Status             string     `json:"status"`
	Mode               string     `json:"mode"`
	CurrentQuestionID  int        `json:"current_question"`
	CurrentGroup       int        `json:"current_group"`
	GroupOrder         []int      `json:"group_order"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	FinishedAt         *time.Time `json:"finished_at,omitempty"`
	TestID             *int       `json:"test_id,omitempty"`
	TestCode           *string    `json:"test_code,omitempty"`
	LastSeen           time.Time  `json:"last_seen"`
}

