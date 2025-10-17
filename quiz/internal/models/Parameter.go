package models

import (
	"encoding/json"
	"io"
)

type Parameter struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ReferenceValues string `json:"reference_values"`
	Order           int    `json:"order"`
}

type ParameterValue struct {
	ParameterID int      `json:"parameter_id"`
	Value1      float64  `json:"value1"`
	Value2      float64  `json:"value2"`
	Value3      *float64 `json:"value3,omitempty"`
}

func (p *ParameterValue) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}
