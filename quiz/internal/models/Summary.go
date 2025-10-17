package models

type QuizSummary struct {
	Questions     int `json:"questions"`
	ActiveSurveys int `json:"active_surveys"`
}
