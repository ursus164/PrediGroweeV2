package models

type Question struct {
	ID            int      `json:"id"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	PredictionAge int      `json:"prediction_age"`
	Case          Case     `json:"case"`
	Correct       *string  `json:"correct"`
	Group         int      `json:"group"`
}

type Case struct {
	ID              int              `json:"id"`
	Code            string           `json:"code"`
	Gender          string           `json:"gender"`
	Age1            int              `json:"age1"`
	Age2            int              `json:"age2"`
	Age3            int              `json:"age3"`
	Parameters      []Parameter      `json:"parameters,omitempty"`
	ParameterValues []ParameterValue `json:"parameters_values,omitempty"`
}

type Parameter struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ReferenceValues string `json:"reference_values"`
	Order           int    `json:"order"`
}

type ParameterValue struct {
	ParameterID int     `json:"parameter_id"`
	Value1      float64 `json:"value1"`
	Value2      float64 `json:"value2"`
	Value3      float64 `json:"value3"`
}

type Option struct {
	ID        int    `json:"id"`
	Option    string `json:"option"`
	Questions *int   `json:"questions,omitempty"`
}
