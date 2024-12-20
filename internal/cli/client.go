package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ASPMClientInterface interface {
	Post(string, interface{}) error
}

type ASPMClient struct {
	serverURL string
	apiKey    string
}

func NewASPMClient(serverUrl string, apiKey string) *ASPMClient {
	return &ASPMClient{
		serverURL: serverUrl,
		apiKey:    apiKey,
	}
}

func (c *ASPMClient) Post(endpoint string, data interface{}) error {
	// Marshal the data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create an HTTP request
	url := fmt.Sprintf("%s%s", c.serverURL, endpoint)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	return nil
}
