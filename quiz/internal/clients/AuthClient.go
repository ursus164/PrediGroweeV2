package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"quiz/internal/models"
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

	req, err := http.NewRequest("POST", c.addr+"/verify", bytes.NewBuffer(jsonPayload))
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
	defer resp.Body.Close()

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

