package gomomo

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TokenResponse represents an OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// AuthService handles authentication with the MTN MoMo API
type AuthService struct {
	client      *Client
	config      *Config
	tokenMutex  sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

// NewAuthService creates a new authentication service
func NewAuthService(client *Client, config *Config) *AuthService {
	return &AuthService{
		client: client,
		config: config,
	}
}

// CreateAPIUser creates a new API user for sandbox environment
func (s *AuthService) CreateAPIUser(ctx context.Context) (string, error) {
	// Only available in sandbox mode
	if s.config.Environment != Sandbox {
		return "", fmt.Errorf("creating API users is only available in sandbox mode")
	}

	apiUserID := uuid.New().String()

	payload := map[string]string{
		"providerCallbackHost": s.config.CallbackHost,
	}

	req := Request{
		Method: http.MethodPost,
		Path:   "/v1_0/apiuser",
		Body:   payload,
		Headers: map[string]string{
			"X-Reference-Id":            apiUserID,
			"Ocp-Apim-Subscription-Key": s.config.SubscriptionKey,
		},
	}

	err := s.client.DoRequest(ctx, req, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create API user: %w", err)
	}

	return apiUserID, nil
}

// CreateAPIKey creates an API key for the given API user
func (s *AuthService) CreateAPIKey(ctx context.Context, apiUserID string) (string, error) {
	// Only available in sandbox mode
	if s.config.Environment != Sandbox {
		return "", fmt.Errorf("creating API keys is only available in sandbox mode")
	}

	var result struct {
		APIKey string `json:"apiKey"`
	}

	req := Request{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("/v1_0/apiuser/%s/apikey", apiUserID),
		Headers: map[string]string{
			"Ocp-Apim-Subscription-Key": s.config.SubscriptionKey,
		},
	}

	err := s.client.DoRequest(ctx, req, &result)
	if err != nil {
		return "", fmt.Errorf("failed to create API key: %w", err)
	}

	return result.APIKey, nil
}

// GetAccessToken fetches a new access token or returns a cached one if still valid
func (s *AuthService) GetAccessToken(ctx context.Context, product string) (string, error) {
	s.tokenMutex.Lock()
	defer s.tokenMutex.Unlock()

	// Check if we have a valid cached token
	if s.accessToken != "" && time.Now().Before(s.tokenExpiry) {
		return s.accessToken, nil
	}

	// Determine which API user and key to use
	apiUser := s.config.APIUser
	apiKey := s.config.APIKey

	// For sandbox, create them if not already set
	if s.config.Environment == Sandbox && (apiUser == "" || apiKey == "") {
		var err error
		apiUser, err = s.CreateAPIUser(ctx)
		if err != nil {
			return "", err
		}

		apiKey, err = s.CreateAPIKey(ctx, apiUser)
		if err != nil {
			return "", err
		}
	}

	// Determine the right path based on product
	tokenPath := ""
	subscriptionKey := ""

	switch product {
	case "collection":
		tokenPath = "/collection/token/"
		subscriptionKey = s.config.SubscriptionKey
	case "disbursement":
		tokenPath = "/disbursement/token/"
		subscriptionKey = s.config.DisbursementKey
	default:
		return "", fmt.Errorf("unknown product: %s", product)
	}

	var tokenResp TokenResponse
	req := Request{
		Method: http.MethodPost,
		Path:   tokenPath,
		Headers: map[string]string{
			"Authorization":             CreateBasicAuthHeader(apiUser, apiKey),
			"Ocp-Apim-Subscription-Key": subscriptionKey,
		},
	}

	err := s.client.DoRequest(ctx, req, &tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to fetch access token: %w", err)
	}

	// Cache the token
	s.accessToken = tokenResp.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // Expire 1 minute early to be safe

	return s.accessToken, nil
}
