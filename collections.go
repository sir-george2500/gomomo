package gomomo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// CollectionService handles MTN MoMo collection operations
type CollectionService struct {
	client      *Client
	config      *Config
	authService *AuthService
}

// NewCollectionService creates a new collection service
func NewCollectionService(client *Client, config *Config, authService *AuthService) *CollectionService {
	return &CollectionService{
		client:      client,
		config:      config,
		authService: authService,
	}
}

type RequestToPayOptions struct {
	IdempotencyKey string // Custom idempotency key (generated if empty)
	ExternalID     string // Custom external ID (generated if empty)
	ReferenceID    string // Custom reference ID (generated if empty)
	Currency       string // Override default currency
	PayerMessage   string // Message to the payer
	PayeeNote      string // Note to the payee
}

// RequestToPay initiates a payment request
func (s *CollectionService) RequestToPay(ctx context.Context, phone string, amount float64, opts *RequestToPayOptions) (string, error) {
	// Format phone number if needed
	phone = formatPhoneNumber(phone)

	// Get access token
	token, err := s.authService.GetAccessToken(ctx, "collection")
	if err != nil {
		return "", fmt.Errorf("error getting access token: %w", err)
	}

	// Use provided options or create defaults
	if opts == nil {
		opts = &RequestToPayOptions{}
	}

	// Generate or use provided reference ID
	referenceID := opts.ReferenceID
	if referenceID == "" {
		referenceID = uuid.New().String()
	}

	// Generate or use provided external ID
	externalID := opts.ExternalID
	if externalID == "" {
		externalID = uuid.New().String()
	}

	// Use provided currency or default
	currency := opts.Currency
	if currency == "" {
		currency = s.config.Currency
	}

	// Create request payload
	payload := RequestToPayPayload{
		Amount:     fmt.Sprintf("%.2f", amount),
		Currency:   currency,
		ExternalID: externalID,
		Payer: PartyInfo{
			PartyIDType: MSISDN,
			PartyID:     phone,
		},
		PayerMessage: defaultIfEmpty(opts.PayerMessage, "Payment request"),
		PayeeNote:    defaultIfEmpty(opts.PayeeNote, "Thank you for your payment"),
	}

	// Create headers
	headers := map[string]string{
		"Authorization":             "Bearer " + token,
		"X-Reference-Id":            referenceID,
		"X-Target-Environment":      s.config.TargetEnvironment,
		"Ocp-Apim-Subscription-Key": s.config.SubscriptionKey,
	}

	// Add idempotency key if provided
	if opts.IdempotencyKey != "" {
		headers["X-Idempotency-Key"] = opts.IdempotencyKey
	}

	// Make the request
	req := Request{
		Method:  http.MethodPost,
		Path:    "/collection/v1_0/requesttopay",
		Body:    payload,
		Headers: headers,
	}

	err = s.client.DoRequest(ctx, req, nil)
	if err != nil {
		return "", fmt.Errorf("error making request-to-pay: %w", err)
	}

	return referenceID, nil
}

// GetTransactionStatus checks the status of a payment request
func (s *CollectionService) GetTransactionStatus(ctx context.Context, referenceID string) (*TransactionStatusResponse, error) {
	// Get access token
	token, err := s.authService.GetAccessToken(ctx, "collection")
	if err != nil {
		return nil, fmt.Errorf("error getting access token: %w", err)
	}

	var result TransactionStatusResponse
	req := Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/collection/v1_0/requesttopay/%s", referenceID),
		Headers: map[string]string{
			"Authorization":             "Bearer " + token,
			"X-Target-Environment":      s.config.TargetEnvironment,
			"Ocp-Apim-Subscription-Key": s.config.SubscriptionKey,
		},
	}

	err = s.client.DoRequest(ctx, req, &result)
	if err != nil {
		return nil, fmt.Errorf("error checking transaction status: %w", err)
	}

	return &result, nil
}

// GetAccountBalance gets the balance of the account
func (s *CollectionService) GetAccountBalance(ctx context.Context) (string, string, error) {
	// Get access token
	token, err := s.authService.GetAccessToken(ctx, "collection")
	if err != nil {
		return "", "", fmt.Errorf("error getting access token: %w", err)
	}

	var result struct {
		AvailableBalance string `json:"availableBalance"`
		Currency         string `json:"currency"`
	}

	req := Request{
		Method: http.MethodGet,
		Path:   "/collection/v1_0/account/balance",
		Headers: map[string]string{
			"Authorization":             "Bearer " + token,
			"X-Target-Environment":      s.config.TargetEnvironment,
			"Ocp-Apim-Subscription-Key": s.config.SubscriptionKey,
		},
	}

	err = s.client.DoRequest(ctx, req, &result)
	if err != nil {
		return "", "", fmt.Errorf("error getting account balance: %w", err)
	}

	return result.AvailableBalance, result.Currency, nil
}

// GetAccountHolderInfo gets information about an account holder
func (s *CollectionService) GetAccountHolderInfo(ctx context.Context, phone string) (*AccountHolderInfo, error) {
	// Format phone number if needed
	phone = formatPhoneNumber(phone)

	// Get access token
	token, err := s.authService.GetAccessToken(ctx, "collection")
	if err != nil {
		return nil, fmt.Errorf("error getting access token: %w", err)
	}

	var result AccountHolderInfo
	req := Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/collection/v1_0/accountholder/MSISDN/%s/basicuserinfo", phone),
		Headers: map[string]string{
			"Authorization":             "Bearer " + token,
			"X-Target-Environment":      s.config.TargetEnvironment,
			"Ocp-Apim-Subscription-Key": s.config.SubscriptionKey,
		},
	}

	err = s.client.DoRequest(ctx, req, &result)
	if err != nil {
		return nil, fmt.Errorf("error getting account holder info: %w", err)
	}

	return &result, nil
}

// Helper to format phone numbers consistently
func formatPhoneNumber(phone string) string {
	// Remove all non-digit characters
	digitsOnly := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// You may want to add specific country code logic here
	// This is a simple example that ensures the number has a country code
	if len(digitsOnly) > 0 && digitsOnly[0] == '0' {
		// Replace leading 0 with country code (e.g., 231 for Liberia)
		digitsOnly = "231" + digitsOnly[1:]
	} else if !strings.HasPrefix(digitsOnly, "231") {
		// Add country code if missing
		digitsOnly = "231" + digitsOnly
	}

	return digitsOnly
}

// Helper for default strings
func defaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
