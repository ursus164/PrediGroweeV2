// Package models defines data structures used across the admin service.
package models

import (
	"encoding/json"
	"io"
)

// User represents a user in the system.
type User struct {
	ID        int      `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Role      UserRole `json:"role"`
	GoogleID  string   `json:"google_id"`
	CreatedAt string   `json:"created_at"`
}

// UserPayload represents the data required to update a user.
type UserPayload struct {
	ID   string   `json:"id"`
	Role UserRole `json:"role"`
}

// ToJSON encodes the user to JSON format.
func (u *User) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(u)
}

// UserStats represents statistics for a user's quiz performance.
type UserStats struct {
	TotalQuestions map[string]int
	CorrectAnswers map[string]int
	Accuracy       map[string]float64
}

// UserDetails aggregates user information, statistics, and survey responses.
type UserDetails struct {
	User            User           `json:"user"`
	Stats           UserStats      `json:"stats"`
	SurveyResponses SurveyResponse `json:"survey"`
}
