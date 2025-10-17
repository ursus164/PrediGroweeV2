package models

type AuthSummary struct {
	Users             int `json:"users"`
	ActiveUsers       int `json:"active_users"`
	Last24hRegistered int `json:"last_registered"`
}
