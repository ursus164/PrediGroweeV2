package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

type AuthClient struct {
	addr   string
	logger *zap.Logger
}

func NewAuthClient(addr string, logger *zap.Logger) *AuthClient {
	return &AuthClient{
		addr:   addr,
		logger: logger,
	}
}

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

	req, err := http.NewRequest("POST", c.addr+"/verify", bytes.NewBuffer(jsonPayload))
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
