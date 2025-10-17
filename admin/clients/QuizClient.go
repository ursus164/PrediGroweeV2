package clients

import (
	"admin/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

type QuizClient interface {
	GetAllQuestions() ([]models.Question, error)
	GetAllParameters() ([]models.Parameter, error)
	UpdateParameter(id string, parameter models.Parameter) error
	DeleteParameter(id string) error
	GetAllOptions() ([]models.Option, error)
	GetQuestion(id string) (models.Question, error)
	UpdateQuestion(id string, question models.Question) error
	CreateParameter(parameter models.Parameter) (models.Parameter, error)
	UpdateOption(id string, option models.Option) error
	CreateOption(option models.Option) (models.Option, error)
	DeleteOption(id string) error
	GetSummary() (models.QuizSummary, error)
	UpdateParametersOrder(order []models.Parameter) error
	GetSettings() ([]models.Settings, error)
	UpdateSettings(settings []models.Settings) error
	ApproveUser(userID int) error
	UnapproveUser(userID int) error
	ListActiveSessions(cutoff int) ([]ActiveSession, error)
	ListSessionsByTestCode(code string) ([]TestSession, error)
}

type QuizRestClient struct {
	addr   string
	apiKey string
	logger *zap.Logger
}

type ActiveSession struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	Status string `json:"status"`
	Mode string `json:"mode"`
	CurrentQuestion int `json:"current_question"`
	CurrentGroup int `json:"current_group"`
	TestID *int `json:"test_id,omitempty"`
	TestCode *string `json:"test_code,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	LastSeen string `json:"last_seen"`
	Accuracy       *float64 `json:"accuracy,omitempty"`
	CorrectAnswers *int     `json:"correct_answers,omitempty"`
	TotalAnswers   *int     `json:"total_answers,omitempty"`
}

type TestSession struct {
	ID              int     `json:"id"`
	UserID          int     `json:"user_id"`
	Status          string  `json:"status"`
	Mode            string  `json:"mode"`
	CurrentQuestion int     `json:"current_question"`
	CurrentGroup    int     `json:"current_group"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	FinishedAt      *string `json:"finished_at,omitempty"`
	TestID          *int    `json:"test_id,omitempty"`
	TestCode        *string `json:"test_code,omitempty"`
}

func NewQuizRestClient(addr string, apiKey string, logger *zap.Logger) *QuizRestClient {
	return &QuizRestClient{
		addr:   addr,
		apiKey: apiKey,
		logger: logger,
	}
}

func (c *QuizRestClient) NewRequestWithAuth(method, path string, body interface{}) (*http.Request, error) {
	var buf *bytes.Buffer
	if body != nil {
		jsonPayload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(jsonPayload)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, c.addr+path, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)

	return req, nil
}

func (c *QuizRestClient) GetAllQuestions() ([]models.Question, error) {
	req, err := c.NewRequestWithAuth("GET", "/questions", nil)
	if err != nil {
		return []models.Question{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []models.Question{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []models.Question{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var questions []models.Question
	err = json.NewDecoder(resp.Body).Decode(&questions)
	return questions, err
}

func (c *QuizRestClient) GetQuestion(id string) (models.Question, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/questions/%s", id), nil)
	if err != nil {
		return models.Question{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.Question{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.Question{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var question models.Question
	err = json.NewDecoder(resp.Body).Decode(&question)
	return question, err
}

func (c *QuizRestClient) GetAllParameters() ([]models.Parameter, error) {
	req, err := c.NewRequestWithAuth("GET", "/parameters", nil)
	if err != nil {
		return []models.Parameter{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []models.Parameter{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []models.Parameter{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var parameters []models.Parameter
	err = json.NewDecoder(resp.Body).Decode(&parameters)
	return parameters, err
}

func (c *QuizRestClient) UpdateParameter(id string, parameter models.Parameter) error {
	req, err := c.NewRequestWithAuth("PATCH", fmt.Sprintf("/parameters/%s", id), parameter)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QuizRestClient) DeleteParameter(id string) error {
	req, err := c.NewRequestWithAuth("DELETE", fmt.Sprintf("/parameters/%s", id), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QuizRestClient) GetAllOptions() ([]models.Option, error) {
	req, err := c.NewRequestWithAuth("GET", "/options", nil)
	if err != nil {
		return []models.Option{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []models.Option{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []models.Option{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var options []models.Option
	err = json.NewDecoder(resp.Body).Decode(&options)
	return options, err
}

func (c *QuizRestClient) UpdateQuestion(id string, question models.Question) error {
	req, err := c.NewRequestWithAuth("PATCH", fmt.Sprintf("/questions/%s", id), question)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QuizRestClient) CreateParameter(parameter models.Parameter) (models.Parameter, error) {
	req, err := c.NewRequestWithAuth("POST", "/parameters", parameter)
	if err != nil {
		return models.Parameter{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.Parameter{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return models.Parameter{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var createdParameter models.Parameter
	if err := json.NewDecoder(resp.Body).Decode(&createdParameter); err != nil {
		return models.Parameter{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return createdParameter, nil
}

func (c *QuizRestClient) UpdateOption(id string, option models.Option) error {
	req, err := c.NewRequestWithAuth("PATCH", fmt.Sprintf("/options/%s", id), option)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QuizRestClient) CreateOption(option models.Option) (models.Option, error) {
	req, err := c.NewRequestWithAuth("POST", "/options", option)
	if err != nil {
		return models.Option{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.Option{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return models.Option{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var createdOption models.Option
	if err := json.NewDecoder(resp.Body).Decode(&createdOption); err != nil {
		return models.Option{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return createdOption, nil
}

func (c *QuizRestClient) DeleteOption(id string) error {
	req, err := c.NewRequestWithAuth("DELETE", fmt.Sprintf("/options/%s", id), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QuizRestClient) GetSummary() (models.QuizSummary, error) {
	req, err := c.NewRequestWithAuth("GET", "/summary", nil)
	if err != nil {
		return models.QuizSummary{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.QuizSummary{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.QuizSummary{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var summary models.QuizSummary
	err = json.NewDecoder(resp.Body).Decode(&summary)
	return summary, err
}

func (c *QuizRestClient) UpdateParametersOrder(order []models.Parameter) error {
	req, err := c.NewRequestWithAuth("PUT", "/parameters/order", order)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QuizRestClient) GetSettings() ([]models.Settings, error) {
	req, err := c.NewRequestWithAuth("GET", "/settings", nil)
	if err != nil {
		return []models.Settings{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []models.Settings{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []models.Settings{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var settings []models.Settings
	err = json.NewDecoder(resp.Body).Decode(&settings)
	return settings, err
}

func (c *QuizRestClient) UpdateSettings(settings []models.Settings) error {
	req, err := c.NewRequestWithAuth("POST", "/settings", settings)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QuizRestClient) ApproveUser(userID int) error {
	body := map[string]int{"user_id": userID}
	req, err := c.NewRequestWithAuth("POST", "/approve", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *QuizRestClient) UnapproveUser(userID int) error {
    body := map[string]int{"user_id": userID}
    req, err := c.NewRequestWithAuth("POST", "/unapprove", body)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    return nil
}


func (c *QuizRestClient) ListActiveSessions(cutoff int) ([]ActiveSession, error) {
	if cutoff <= 0 {
		cutoff = 5
	}
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/sessions/active?cutoff=%d", cutoff), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var out []ActiveSession
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *QuizRestClient) ListSessionsByTestCode(code string) ([]TestSession, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/tests/%s/sessions", code), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var out []TestSession
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode body: %w", err)
	}
	return out, nil
}
