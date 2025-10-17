package models

import (
	"encoding/json"
	"io"
)

type GoogleTokenPayload struct {
	Token string `json:"access_token"`
}

func (g *GoogleTokenPayload) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(g)
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
}

func (g *GoogleUserInfo) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(g)
}
