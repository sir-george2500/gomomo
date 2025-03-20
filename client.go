package gomomo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles HTTP communication with the MTN MoMo API
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new MTN MoMo API client
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request represents an HTTP request to the API
type Request struct {
	Method      string
	Path        string
	Body        interface{}
	Headers     map[string]string
	QueryParams map[string]string
}

// DoRequest performs an HTTP request and decodes the response
func (c *Client) DoRequest(ctx context.Context, req Request, result interface{}) error {
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	url := fmt.Sprintf("https://%s%s", c.config.Host, req.Path)
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")

	// Add request-specific headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add query parameters
	if len(req.QueryParams) > 0 {
		q := httpReq.URL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Only try to decode if we have a result pointer and the response isn't empty
	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
	}

	return nil
}

// CreateBasicAuthHeader generates a Basic Auth header from API user and key
func CreateBasicAuthHeader(apiUser, apiKey string) string {
	auth := fmt.Sprintf("%s:%s", apiUser, apiKey)
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
