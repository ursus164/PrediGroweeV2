package models

import (
	"encoding/json"
	"io"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type UserData struct {
	UserID int      `json:"user_id"`
	Role   UserRole `json:"role"`
}

func (u *UserData) FromJSON(ioReader io.Reader) error {
	return json.NewDecoder(ioReader).Decode(u)
}
