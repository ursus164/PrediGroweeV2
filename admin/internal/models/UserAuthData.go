package models

import (
	"encoding/json"
	"io"
)

// UserRole represents a user's role in the system.
type UserRole string

const (
	// RoleAdmin represents an administrator user.
	RoleAdmin UserRole = "admin"
	// RoleUser represents a regular user.
	RoleUser UserRole = "user"
	// RoleTeacher represents a teacher user.
	RoleTeacher UserRole = "teacher"
)

// UserAuthData contains authentication data for a user.
type UserAuthData struct {
	UserID int      `json:"user_id"`
	Role   UserRole `json:"role"`
}

// FromJSON decodes JSON data into UserAuthData.
func (u *UserAuthData) FromJSON(ioReader io.Reader) error {
	return json.NewDecoder(ioReader).Decode(u)
}
