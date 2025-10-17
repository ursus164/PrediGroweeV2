package models

type Summary struct {
	QuizSummary  QuizSummary  `json:"quiz_summary"`
	StatsSummary StatsSummary `json:"stats_summary"`
	AuthSummary  AuthSummary  `json:"auth_summary"`
}
type QuizSummary struct {
	Questions     int `json:"questions"`
	ActiveSurveys int `json:"active_surveys"`
}

type StatsSummary struct {
	QuizSessions   int `json:"quiz_sessions"`
	TotalResponses int `json:"total_responses"`
	TotalCorrect   int `json:"total_correct"`
}

type AuthSummary struct {
	Users             int `json:"users"`
	ActiveUsers       int `json:"active_users"`
	Last24hRegistered int `json:"last_registered"`
}
