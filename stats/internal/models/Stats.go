package models

import (
	"encoding/json"
	"io"
	"time"
)

// QuizMode represents the type of quiz mode.
type QuizMode = string

// Quiz mode constants
const (
	QuizModeEducational QuizMode = "educational"
	QuizModeClassic     QuizMode = "classic"
	QuizModeLimitedTime QuizMode = "time_limited"
)

// UserStats represents statistics for a user across different quiz modes.
type UserStats struct {
	TotalQuestions map[QuizMode]int
	CorrectAnswers map[QuizMode]int
	Accuracy       map[QuizMode]float64
}

// QuestionStat represents statistics for a single question.
type QuestionStat struct {
	QuestionID int
	Answer     string
	IsCorrect  bool
}

// QuizStats represents detailed statistics for a quiz session.
type QuizStats struct {
	SessionID      int            `json:"session_id"`
	Mode           QuizMode       `json:"mode"`
	TotalQuestions int            `json:"total_questions"`
	CorrectAnswers int            `json:"correct_answers"`
	Accuracy       float64        `json:"accuracy"`
	Questions      []QuestionStat `json:"questions"`
	StartTime      *time.Time     `json:"start_time"`
}

// UserQuizStats represents aggregated quiz statistics for a user.
type UserQuizStats struct {
	UserID         int    `json:"user_id"`
	TotalAnswers   int    `json:"total_answers"`
	CorrectAnswers int    `json:"correct_answers"`
	Experience     string `json:"experience"`
	Education      string `json:"education"`
}

// ToJSON encodes QuizStats to JSON.
func (s *QuizStats) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(s)
}

// ToJSON encodes UserStats to JSON.
func (u *UserStats) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(u)
}

// QuestionAllStats represents aggregated statistics for a specific question.
type QuestionAllStats struct {
	QuestionID int    `json:"question_id"`
	CaseCode   string `json:"case_id"`
	Total      int    `json:"total"`
	Correct    int    `json:"correct"`
}

// ActivityStats represents daily activity statistics.
type ActivityStats struct {
	Date    time.Time `json:"date"`
	Total   int       `json:"total"`
	Correct int       `json:"correct"`
}

// SurveyGroupedStats represents statistics grouped by survey field.
type SurveyGroupedStats struct {
	Group    string  `json:"group"`
	Value    string  `json:"value"`
	Total    int     `json:"total"`
	Correct  int     `json:"correct"`
	Accuracy float64 `json:"accuracy"`
}
