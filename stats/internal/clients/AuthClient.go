// Package clients provides HTTP client implementations for communicating with external services.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"stats/internal/models"
)

// AuthClient is an HTTP client for authenticating requests with the auth service.
type AuthClient struct {
	addr   string
	logger *zap.Logger
}

// NewAuthClient creates a new AuthClient instance.
func NewAuthClient(addr string, logger *zap.Logger) *AuthClient {
	return &AuthClient{
		addr:   addr,
		logger: logger,
	}
}

// VerifyAuthToken verifies an authentication token and returns the associated user data.
func (c *AuthClient) VerifyAuthToken(token string) (models.UserData, error) {
	body := struct {
		AuthToken string `json:"token"`
	}{
		AuthToken: token,
	}

	jsonPayload, err := json.Marshal(body)
	if err != nil {
		return models.UserData{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", c.addr+"/verify", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return models.UserData{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.UserData{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("failed to close response body", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return models.UserData{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var userDataResponse models.UserData
	err = json.NewDecoder(resp.Body).Decode(&userDataResponse)
	if err != nil {
		return models.UserData{}, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("response", zap.Any("response", userDataResponse))

	return userDataResponse, nil
}
