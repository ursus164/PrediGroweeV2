package models

import "time"

type UserSession struct {
	UserID     int
	Token      string
	SessionID  string
	Expiration time.Time
}
