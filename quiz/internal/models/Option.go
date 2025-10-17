package models

type Option struct {
	ID        int    `json:"id"`
	Option    string `json:"option"`
	Questions *int   `json:"questions,omitempty"`
}
