// Package clients contains external service clients used by the Images service.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// AuthClient provides methods to call the Auth service.
type AuthClient struct {
	addr   string
	logger *zap.Logger
}

// NewAuthClient constructs an AuthClient.
func NewAuthClient(addr string, logger *zap.Logger) *AuthClient {
	return &AuthClient{
		addr:   addr,
		logger: logger,
	}
}

// VerifyAuthToken validates the provided token against the Auth service.
func (c *AuthClient) VerifyAuthToken(token string) error {
	body := struct {
		AuthToken string `json:"token"`
	}{
		AuthToken: token,
	}

	jsonPayload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.addr+"/verify", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.logger.Warn("failed to close auth response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
