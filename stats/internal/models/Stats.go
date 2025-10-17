package models

import (
	"encoding/json"
	"io"
	"time"
)

type QuizMode = string

const (
	QuizModeEducational QuizMode = "educational"
	QuizModeClassic     QuizMode = "classic"
	QuizModeLimitedTime QuizMode = "time_limited"
)

type UserStats struct {
	TotalQuestions map[QuizMode]int
	CorrectAnswers map[QuizMode]int
	Accuracy       map[QuizMode]float64
}
type QuestionStat struct {
	QuestionID int
	Answer     string
	IsCorrect  bool
}

type QuizStats struct {
	SessionID      int            `json:"session_id"`
	Mode           QuizMode       `json:"mode"`
	TotalQuestions int            `json:"total_questions"`
	CorrectAnswers int            `json:"correct_answers"`
	Accuracy       float64        `json:"accuracy"`
	Questions      []QuestionStat `json:"questions"`
	StartTime      *time.Time     `json:"start_time"`
}

type UserQuizStats struct {
	UserID         int    `json:"user_id"`
	TotalAnswers   int    `json:"total_answers"`
	CorrectAnswers int    `json:"correct_answers"`
	Experience     string `json:"experience"`
	Education      string `json:"education"`
}

func (s *QuizStats) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(s)
}

func (u *UserStats) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(u)
}

type QuestionAllStats struct {
	QuestionID int    `json:"question_id"`
	CaseCode   string `json:"case_id"`
	Total      int    `json:"total"`
	Correct    int    `json:"correct"`
}

type ActivityStats struct {
	Date    time.Time `json:"date"`
	Total   int       `json:"total"`
	Correct int       `json:"correct"`
}
type SurveyGroupedStats struct {
	Group    string  `json:"group"`
	Value    string  `json:"value"`
	Total    int     `json:"total"`
	Correct  int     `json:"correct"`
	Accuracy float64 `json:"accuracy"`
}
