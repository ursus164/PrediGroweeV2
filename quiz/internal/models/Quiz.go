package models

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"io"
)

type QuizMode = string
type QuizStatus = string

const (
	QuizModeEducational QuizMode = "educational"
	QuizModeClassic     QuizMode = "classic"
	QuizModeLimitedTime QuizMode = "limited_time"
)
const (
	QuizStatusNotStarted QuizStatus = "not_started"
	QuizStatusInProgress QuizStatus = "in_progress"
	QuizStatusFinished   QuizStatus = "finished"
)

type StartQuizPayload struct {
	Mode         QuizMode `json:"mode" ,validate:"required,oneof=educational classic limited_time"`
	ScreenWidth  int      `json:"screen_width" ,validate:"required"`
	ScreenHeight int      `json:"screen_height" ,validate:"required"`
	TestCode    string `json:"test_code,omitempty"`
}

func (p *StartQuizPayload) Validate() error {
	return validator.New().Struct(p)
}
func (p *StartQuizPayload) FromJSON(ioReader io.Reader) error {
	return json.NewDecoder(ioReader).Decode(p)
}
