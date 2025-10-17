package clients

import (
	"admin/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

type AuthClient interface {
	VerifyAuthToken(token string) (models.UserAuthData, error)
	GetUsers() ([]models.User, error)
	UpdateUser(user models.UserPayload) error
	GetUser(id string) (models.User, error)
	DeleteUser(id string) error
	GetSummary() (models.AuthSummary, error)
}

type RestAuthClient struct {
	addr   string
	apiKey string
	logger *zap.Logger
}

func NewRestAuthClient(addr string, apiKey string, logger *zap.Logger) *RestAuthClient {
	return &RestAuthClient{
		addr:   addr,
		apiKey: apiKey,
		logger: logger,
	}
}

func (c *RestAuthClient) NewRequestWithAuth(method, path string, body interface{}) (*http.Request, error) {
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		c.logger.Error("unexpected status code", zap.Error(err), zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
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
	defer resp.Body.Close()
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
