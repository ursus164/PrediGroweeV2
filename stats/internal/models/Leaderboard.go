// Package models provides data structures for the stats service.
package models

// LeaderboardRow represents a single entry in the quiz leaderboard.
type LeaderboardRow struct {
	UserID         int     `json:"user_id"`
	Education      *string `json:"education,omitempty"`
	Country        *string `json:"country,omitempty"`
	TotalAnswers   int     `json:"total_answers"`
	CorrectAnswers int     `json:"correct_answers"`
	Accuracy       float64 `json:"accuracy"`
}

