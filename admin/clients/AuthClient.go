// Package clients provides HTTP client implementations for interacting with various microservices.
package clients

import (
	"admin/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// AuthClient defines the interface for authentication service operations.
type AuthClient interface {
	VerifyAuthToken(token string) (models.UserAuthData, error)
	GetUsers() ([]models.User, error)
	UpdateUser(user models.UserPayload) error
	GetUser(id string) (models.User, error)
	DeleteUser(id string) error
	GetSummary() (models.AuthSummary, error)
}

// RestAuthClient is an HTTP client for the authentication service.
type RestAuthClient struct {
	addr   string
	apiKey string
	logger *zap.Logger
}

// NewRestAuthClient creates a new instance of RestAuthClient.
func NewRestAuthClient(addr string, apiKey string, logger *zap.Logger) *RestAuthClient {
	return &RestAuthClient{
		addr:   addr,
		apiKey: apiKey,
		logger: logger,
	}
}

// NewRequestWithAuth creates a new HTTP request with authentication headers.
func (c *RestAuthClient) NewRequestWithAuth(method, path string, body interface{}) (*http.Request, error) {
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

// VerifyAuthToken verifies an authentication token and returns user data.
func (c *RestAuthClient) VerifyAuthToken(token string) (models.UserAuthData, error) {

	req, err := c.NewRequestWithAuth("POST", "/verify", nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return models.UserAuthData{}, fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.UserAuthData{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return models.UserAuthData{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var userDataResponse models.UserAuthData
	err = json.NewDecoder(resp.Body).Decode(&userDataResponse)
	if err != nil {
		return models.UserAuthData{}, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("response", zap.Any("response", userDataResponse))

	return userDataResponse, nil
}

// GetUsers retrieves all users from the authentication service.
func (c *RestAuthClient) GetUsers() ([]models.User, error) {
	req, err := c.NewRequestWithAuth("GET", "/users", nil)
	if err != nil {
		c.logger.Error("failed to create request", zap.Error(err))
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("failed to send request", zap.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var users []models.User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		c.logger.Error("failed to decode response", zap.Error(err))
		return nil, err
	}

	c.logger.Info("response", zap.Any("response", users))
	return users, nil
}

// UpdateUser updates a user's information.
func (c *RestAuthClient) UpdateUser(user models.UserPayload) error {
	req, err := c.NewRequestWithAuth("PATCH", "/users/"+user.ID, user)
	if err != nil {
		c.logger.Error("failed to create request", zap.Error(err))
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("failed to send request", zap.Error(err))
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// GetUser retrieves a specific user by ID.
func (c *RestAuthClient) GetUser(id string) (models.User, error) {
	req, err := c.NewRequestWithAuth("GET", "/users/"+id, nil)
	if err != nil {
		c.logger.Error("failed to create request", zap.Error(err))
		return models.User{}, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("failed to send request", zap.Error(err))
		return models.User{}, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return models.User{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var user models.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		c.logger.Error("failed to decode response", zap.Error(err))
		return models.User{}, err
	}
	return user, nil
}

// DeleteUser deletes a user by ID.
func (c *RestAuthClient) DeleteUser(id string) error {
	req, err := c.NewRequestWithAuth("DELETE", "/users/"+id, nil)
	if err != nil {
		c.logger.Error("failed to create request", zap.Error(err))
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("failed to send request", zap.Error(err))
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusNoContent {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// GetSummary retrieves authentication service summary statistics.
func (c *RestAuthClient) GetSummary() (models.AuthSummary, error) {
	req, err := c.NewRequestWithAuth("GET", "/summary", nil)
	if err != nil {
		c.logger.Error("failed to create request", zap.Error(err))
		return models.AuthSummary{}, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("failed to send request", zap.Error(err))
		return models.AuthSummary{}, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("failed to close response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return models.AuthSummary{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var summary models.AuthSummary
	err = json.NewDecoder(resp.Body).Decode(&summary)
	if err != nil {
		c.logger.Error("failed to decode response", zap.Error(err))
		return models.AuthSummary{}, err
	}
	return summary, nil
}
