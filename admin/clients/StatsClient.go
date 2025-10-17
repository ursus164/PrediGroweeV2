package clients

import (
	"admin/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type StatsClient interface {
	GetUserStats(userID string) (models.UserStats, error)
	GetAllResponses() ([]models.QuestionResponse, error)
	GetStatsForQuestion(id string) (models.QuestionStats, error)
	GetStatsForAllQuestions() ([]models.QuestionStats, error)
	GetActivityStats() ([]models.ActivityStats, error)
	GetSummary() (models.StatsSummary, error)
	GetSurvey(id string) (models.SurveyResponse, error)
	GetAllSurveys() ([]models.SurveyResponse, error)
	GetStatsGroupedBySurvey(groupBy string) ([]models.SurveyGroupedStats, error)
	DeleteResponse(id string) error
	DeleteUserResponses(id string) error
	GetAllUsersStats() ([]models.UserQuizStats, error)
	GetSessionsAccuracy(sessionIDs []int) ([]models.SessionAccuracy, error)
}

type StatsRestClient struct {
	addr   string
	apiKey string
	logger *zap.Logger
}

func NewStatsRestClient(addr string, apiKey string, logger *zap.Logger) *StatsRestClient {
	return &StatsRestClient{
		addr:   addr,
		apiKey: apiKey,
		logger: logger,
	}
}
func (c *StatsRestClient) NewRequestWithAuth(method, path string, body interface{}) (*http.Request, error) {
	jsonPayload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(method, c.addr+path, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)

	return req, nil
}

func (c *StatsRestClient) MakeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return resp, nil
}

func (c *StatsRestClient) GetUserStats(userID string) (models.UserStats, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/users/%s", userID), nil)
	if err != nil {
		return models.UserStats{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.UserStats{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.UserStats{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var userStats models.UserStats
	err = json.NewDecoder(resp.Body).Decode(&userStats)
	if err != nil {
		return models.UserStats{}, fmt.Errorf("failed to decode response body: %w", err)
	}
	return userStats, nil
}

func (c *StatsRestClient) GetAllResponses() ([]models.QuestionResponse, error) {
	req, err := c.NewRequestWithAuth("GET", "/responses", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []models.QuestionResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var responses []models.QuestionResponse
	err = json.NewDecoder(resp.Body).Decode(&responses)
	if err != nil {
		return []models.QuestionResponse{}, fmt.Errorf("failed to decode response body: %w", err)
	}
	return responses, nil
}
func (c *StatsRestClient) GetStatsForQuestion(id string) (models.QuestionStats, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/questions/%s/stats", id), nil)
	if err != nil {
		return models.QuestionStats{}, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return models.QuestionStats{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.QuestionStats{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var stats models.QuestionStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return models.QuestionStats{}, fmt.Errorf("failed to decode response body: %w", err)
	}
	return stats, nil
}
func (c *StatsRestClient) GetStatsForAllQuestions() ([]models.QuestionStats, error) {
	req, err := c.NewRequestWithAuth("GET", "/questions/-/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []models.QuestionStats{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var stats []models.QuestionStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return []models.QuestionStats{}, fmt.Errorf("failed to decode response body: %w", err)
	}
	return stats, nil
}

func (c *StatsRestClient) GetActivityStats() ([]models.ActivityStats, error) {
	req, err := c.NewRequestWithAuth("GET", "/activity", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []models.ActivityStats{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var stats []models.ActivityStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return []models.ActivityStats{}, fmt.Errorf("failed to decode response body: %w", err)
	}
	return stats, nil
}
func (c *StatsRestClient) GetSummary() (models.StatsSummary, error) {
	req, err := c.NewRequestWithAuth("GET", "/summary", nil)
	if err != nil {
		return models.StatsSummary{}, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return models.StatsSummary{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.StatsSummary{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var summary models.StatsSummary
	err = json.NewDecoder(resp.Body).Decode(&summary)
	if err != nil {
		return models.StatsSummary{}, fmt.Errorf("failed to decode response body: %w", err)
	}
	return summary, nil
}

func (c *StatsRestClient) GetSurvey(id string) (models.SurveyResponse, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/surveys/users/%s", id), nil)
	if err != nil {
		return models.SurveyResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return models.SurveyResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return models.SurveyResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var survey models.SurveyResponse
	err = json.NewDecoder(resp.Body).Decode(&survey)
	if err != nil {
		return models.SurveyResponse{}, fmt.Errorf("failed to decode response body: %w", err)
	}
	return survey, nil
}
func (c *StatsRestClient) GetAllSurveys() ([]models.SurveyResponse, error) {
	req, err := c.NewRequestWithAuth("GET", "/surveys/users/-", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []models.SurveyResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var surveys []models.SurveyResponse
	err = json.NewDecoder(resp.Body).Decode(&surveys)
	if err != nil {
		return []models.SurveyResponse{}, fmt.Errorf("failed to decode response body: %w", err)
	}
	return surveys, nil
}

func (c *StatsRestClient) GetStatsGroupedBySurvey(groupBy string) ([]models.SurveyGroupedStats, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/grouped?groupBy=%s", groupBy), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var stats []models.SurveyGroupedStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	return stats, err
}

func (c *StatsRestClient) DeleteResponse(id string) error {
	req, err := c.NewRequestWithAuth("DELETE", fmt.Sprintf("/responses/%s", id), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *StatsRestClient) DeleteUserResponses(id string) error {
	req, err := c.NewRequestWithAuth("DELETE", fmt.Sprintf("/users/%s/responses", id), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
func (c *StatsRestClient) GetAllUsersStats() ([]models.UserQuizStats, error) {
	req, err := c.NewRequestWithAuth("GET", "/users/stats", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats []models.UserQuizStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	return stats, err
}
func (c *StatsRestClient) GetSessionsAccuracy(sessionIDs []int) ([]models.SessionAccuracy, error) {
	if len(sessionIDs) == 0 {
		return []models.SessionAccuracy{}, nil
	}
	b := strings.Builder{}
	for i, id := range sessionIDs {
		if i > 0 { b.WriteByte(',') }
		b.WriteString(fmt.Sprintf("%d", id))
	}
	path := fmt.Sprintf("/sessions/accuracy?ids=%s", b.String())

	req, err := c.NewRequestWithAuth("GET", path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var out []models.SessionAccuracy
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
