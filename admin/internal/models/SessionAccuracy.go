package models

// SessionAccuracy represents accuracy statistics for a quiz session.
type SessionAccuracy struct {
	SessionID int     `json:"session_id"`
	Correct   int     `json:"correct"`
	Total     int     `json:"total"`
	Accuracy  float64 `json:"accuracy"`
}
