package models

import "time"

type CaseReport struct {
	ID          int       `json:"id"`
	CaseID      int       `json:"case_id"`
	CaseCode    string    `json:"case_code"`
	UserID      int       `json:"user_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	AdminNote          *string    `json:"admin_note,omitempty"`
	AdminNoteUpdatedAt *time.Time `json:"admin_note_updated_at,omitempty"`
	AdminNoteUpdatedBy *int       `json:"admin_note_updated_by,omitempty"`
}


