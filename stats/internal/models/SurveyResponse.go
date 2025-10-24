package models

import (
	"encoding/json"
	"io"
)

// SurveyResponse represents a user's survey response with demographic information.
type SurveyResponse struct {
	UserID          int    `json:"user_id"`
	Gender          string `json:"gender"`
	Age             string `json:"age"`
	VisionDefect    string `json:"vision_defect"`
	Education       string `json:"education"`
	Experience      string `json:"experience"`
	Country         string `json:"country"`
	Name            string `json:"name"`
	Surname         string `json:"surname"`
	Acknowledgments string `json:"acknowledgments"`
}

// FromJSON decodes a SurveyResponse from JSON.
func (s *SurveyResponse) FromJSON(reader io.Reader) error {
	return json.NewDecoder(reader).Decode(s)
}
