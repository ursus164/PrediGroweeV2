package models

// StatsSummary represents a summary of overall statistics.
type StatsSummary struct {
	QuizSessions   int `json:"quiz_sessions"`
	TotalResponses int `json:"total_responses"`
	TotalCorrect   int `json:"total_correct"`
}
