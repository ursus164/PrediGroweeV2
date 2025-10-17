package models

import (
	"encoding/json"
	"io"
)

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleUser    UserRole = "user"
	RoleTeacher UserRole = "teacher"
)

type UserAuthData struct {
	UserID int      `json:"user_id"`
	Role   UserRole `json:"role"`
}

func (u *UserAuthData) FromJSON(ioReader io.Reader) error {
	return json.NewDecoder(ioReader).Decode(u)
}
