package models

// QuestionsGroup represents a group of questions
type QuestionsGroup struct {
	ID           int   `json:"id"`
	QuestionsIDs []int `json:"questions"`
}
