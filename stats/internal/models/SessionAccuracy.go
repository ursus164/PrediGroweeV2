package models

type SessionAccuracy struct {
	SessionID int     `json:"session_id"`
	Correct   int     `json:"correct"`
	Total     int     `json:"total"`
	Accuracy  float64 `json:"accuracy"`
}

