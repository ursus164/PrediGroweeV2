package models

import (
	"encoding/json"
	"io"
)

// UserRole represents the role of a user.
type UserRole string

// User role constants
const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// UserData represents authenticated user information.
type UserData struct {
	UserID int      `json:"user_id"`
	Role   UserRole `json:"role"`
}

// FromJSON decodes UserData from JSON.
func (u *UserData) FromJSON(ioReader io.Reader) error {
	return json.NewDecoder(ioReader).Decode(u)
}
