package models

import "time"

type DifficultyLevel string

const (
	DifficultyEasy DifficultyLevel = "easy"
	DifficultyHard DifficultyLevel = "hard"
)

type QuestionDifficultyVote struct {
	QuestionID int             `json:"question_id" db:"question_id"`
	UserID     int             `json:"user_id" db:"user_id"`
	Difficulty DifficultyLevel `json:"difficulty" db:"difficulty"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

type QuestionDifficultySummary struct {
	QuestionID int     `json:"question_id" db:"question_id"`
	TotalVotes int     `json:"total_votes" db:"total_votes"`
	HardVotes  int     `json:"hard_votes" db:"hard_votes"`
	EasyVotes  int     `json:"easy_votes" db:"easy_votes"`
	HardPct    float64 `json:"hard_pct" db:"hard_pct"`
}

