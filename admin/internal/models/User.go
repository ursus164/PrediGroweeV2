package models

import (
	"encoding/json"
	"io"
)

type User struct {
	ID        int      `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Role      UserRole `json:"role"`
	GoogleID  string   `json:"google_id"`
	CreatedAt string   `json:"created_at"`
}

type UserPayload struct {
	ID   string   `json:"id"`
	Role UserRole `json:"role"`
}

func (u *User) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(u)
}

type UserStats struct {
	TotalQuestions map[string]int
	CorrectAnswers map[string]int
	Accuracy       map[string]float64
}

type UserDetails struct {
	User            User           `json:"user"`
	Stats           UserStats      `json:"stats"`
	SurveyResponses SurveyResponse `json:"survey"`
}
