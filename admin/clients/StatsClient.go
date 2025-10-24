package clients

import (
	"admin/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// StatsClient defines the interface for stats service operations.
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

// StatsRestClient is an HTTP client for the stats service.
type StatsRestClient struct {
	addr   string
	apiKey string
	logger *zap.Logger
}

// NewStatsRestClient creates a new instance of StatsRestClient.
func NewStatsRestClient(addr string, apiKey string, logger *zap.Logger) *StatsRestClient {
	return &StatsRestClient{
		addr:   addr,
		apiKey: apiKey,
		logger: logger,
	}
}

// NewRequestWithAuth creates a new HTTP request with authentication headers.
func (c *StatsRestClient) NewRequestWithAuth(method, path string, body interface{}) (*http.Request, error) {
	jsonPayload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, c.addr+path, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)

	return req, nil
}

// MakeRequest executes an HTTP request.
func (c *StatsRestClient) MakeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	return resp, nil
}

// GetUserStats retrieves statistics for a specific user.
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
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

// GetAllResponses retrieves all question responses.
func (c *StatsRestClient) GetAllResponses() ([]models.QuestionResponse, error) {
	req, err := c.NewRequestWithAuth("GET", "/responses", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
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

// GetStatsForQuestion retrieves statistics for a specific question.
func (c *StatsRestClient) GetStatsForQuestion(id string) (models.QuestionStats, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/questions/%s/stats", id), nil)
	if err != nil {
		return models.QuestionStats{}, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return models.QuestionStats{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
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

// GetStatsForAllQuestions retrieves statistics for all questions.
func (c *StatsRestClient) GetStatsForAllQuestions() ([]models.QuestionStats, error) {
	req, err := c.NewRequestWithAuth("GET", "/questions/-/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
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

// GetActivityStats retrieves activity statistics.
func (c *StatsRestClient) GetActivityStats() ([]models.ActivityStats, error) {
	req, err := c.NewRequestWithAuth("GET", "/activity", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
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

// GetSummary retrieves stats service summary statistics.
func (c *StatsRestClient) GetSummary() (models.StatsSummary, error) {
	req, err := c.NewRequestWithAuth("GET", "/summary", nil)
	if err != nil {
		return models.StatsSummary{}, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return models.StatsSummary{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
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

// GetSurvey retrieves survey data for a specific user.
func (c *StatsRestClient) GetSurvey(id string) (models.SurveyResponse, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/surveys/users/%s", id), nil)
	if err != nil {
		return models.SurveyResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return models.SurveyResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
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

// GetAllSurveys retrieves all survey responses.
func (c *StatsRestClient) GetAllSurveys() ([]models.SurveyResponse, error) {
	req, err := c.NewRequestWithAuth("GET", "/surveys/users/-", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
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

// GetStatsGroupedBySurvey retrieves statistics grouped by survey parameter.
func (c *StatsRestClient) GetStatsGroupedBySurvey(groupBy string) ([]models.SurveyGroupedStats, error) {
	req, err := c.NewRequestWithAuth("GET", fmt.Sprintf("/grouped?groupBy=%s", groupBy), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var stats []models.SurveyGroupedStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	return stats, err
}

// DeleteResponse deletes a specific response by ID.
func (c *StatsRestClient) DeleteResponse(id string) error {
	req, err := c.NewRequestWithAuth("DELETE", fmt.Sprintf("/responses/%s", id), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// DeleteUserResponses deletes all responses for a specific user.
func (c *StatsRestClient) DeleteUserResponses(id string) error {
	req, err := c.NewRequestWithAuth("DELETE", fmt.Sprintf("/users/%s/responses", id), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// GetAllUsersStats retrieves quiz statistics for all users.
func (c *StatsRestClient) GetAllUsersStats() ([]models.UserQuizStats, error) {
	req, err := c.NewRequestWithAuth("GET", "/users/stats", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()

	var stats []models.UserQuizStats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	return stats, err
}

// GetSessionsAccuracy retrieves accuracy statistics for specified session IDs.
func (c *StatsRestClient) GetSessionsAccuracy(sessionIDs []int) ([]models.SessionAccuracy, error) {
	if len(sessionIDs) == 0 {
		return []models.SessionAccuracy{}, nil
	}
	b := strings.Builder{}
	for i, id := range sessionIDs {
		if i > 0 {
			b.WriteByte(',')
		}
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var out []models.SessionAccuracy
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
