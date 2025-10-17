package models

import "time"

type FavoriteCase struct {
	CaseID           int               `json:"case_id"`
	CaseCode         string            `json:"case_code"`
	QuestionID       *int              `json:"question_id,omitempty"`
	Correct          *string           `json:"correct,omitempty"`
	Gender           string            `json:"gender"`
	Age1             int               `json:"age1"`
	Age2             int               `json:"age2"`
	Age3             *int              `json:"age3,omitempty"`
	Parameters       []Parameter       `json:"parameters"`
	ParameterValues  []ParameterValue  `json:"parametersValues"`
	CreatedAt        time.Time         `json:"created_at"`
	Note          *string    `json:"note,omitempty"`
	NoteUpdatedAt *time.Time `json:"note_updated_at,omitempty"`
}

