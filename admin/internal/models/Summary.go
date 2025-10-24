package models

// Summary aggregates summaries from different services.
type Summary struct {
	QuizSummary  QuizSummary  `json:"quiz_summary"`
	StatsSummary StatsSummary `json:"stats_summary"`
	AuthSummary  AuthSummary  `json:"auth_summary"`
}

// QuizSummary represents summary statistics from the quiz service.
type QuizSummary struct {
	Questions     int `json:"questions"`
	ActiveSurveys int `json:"active_surveys"`
}

// StatsSummary represents summary statistics from the stats service.
type StatsSummary struct {
	QuizSessions   int `json:"quiz_sessions"`
	TotalResponses int `json:"total_responses"`
	TotalCorrect   int `json:"total_correct"`
}

// AuthSummary represents summary statistics from the auth service.
type AuthSummary struct {
	Users             int `json:"users"`
	ActiveUsers       int `json:"active_users"`
	Last24hRegistered int `json:"last_registered"`
}
