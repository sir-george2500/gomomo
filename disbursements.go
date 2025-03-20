package gomomo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// DisbursementService handles MTN MoMo disbursement operations
type DisbursementService struct {
	client      *Client
	config      *Config
	authService *AuthService
}

// NewDisbursementService creates a new disbursement service
func NewDisbursementService(client *Client, config *Config, authService *AuthService) *DisbursementService {
	return &DisbursementService{
		client:      client,
		config:      config,
		authService: authService,
	}
}

// TransferOptions contains optional parameters for transfers
type TransferOptions struct {
	IdempotencyKey string // Custom idempotency key (generated if empty)
	ExternalID     string // Custom external ID (generated if empty)
	ReferenceID    string // Custom reference ID (generated if empty)
	Currency       string // Override default currency
	PayerMessage   string // Message from the payer
	PayeeNote      string // Note to the payee
}

// Transfer initiates a transfer to a mobile money account
func (s *DisbursementService) Transfer(ctx context.Context, phone string, amount float64, opts *TransferOptions) (string, error) {
	// Format phone number if needed
	phone = formatPhoneNumber(phone)

	// Get access token
	token, err := s.authService.GetAccessToken(ctx, "disbursement")
	if err != nil {
		return "", fmt.Errorf("error getting access token: %w", err)
	}

	// Use provided options or create defaults
	if opts == nil {
		opts = &TransferOptions{}
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
	payload := TransferPayload{
		Amount:     fmt.Sprintf("%.2f", amount),
		Currency:   currency,
		ExternalID: externalID,
		Payee: PartyInfo{
			PartyIDType: MSISDN,
			PartyID:     phone,
		},
		PayerMessage: defaultIfEmpty(opts.PayerMessage, "Disbursement payment"),
		PayeeNote:    defaultIfEmpty(opts.PayeeNote, "Funds received"),
	}

	// Create headers
	headers := map[string]string{
		"Authorization":             "Bearer " + token,
		"X-Reference-Id":            referenceID,
		"X-Target-Environment":      s.config.TargetEnvironment,
		"Ocp-Apim-Subscription-Key": s.config.DisbursementKey,
	}

	// Add idempotency key if provided
	if opts.IdempotencyKey != "" {
		headers["X-Idempotency-Key"] = opts.IdempotencyKey
	}

	// Make the request
	req := Request{
		Method:  http.MethodPost,
		Path:    "/disbursement/v1_0/transfer",
		Body:    payload,
		Headers: headers,
	}

	err = s.client.DoRequest(ctx, req, nil)
	if err != nil {
		return "", fmt.Errorf("error making transfer: %w", err)
	}

	return referenceID, nil
}

// GetTransferStatus checks the status of a transfer
func (s *DisbursementService) GetTransferStatus(ctx context.Context, referenceID string) (*TransactionStatusResponse, error) {
	// Get access token
	token, err := s.authService.GetAccessToken(ctx, "disbursement")
	if err != nil {
		return nil, fmt.Errorf("error getting access token: %w", err)
	}

	var result TransactionStatusResponse
	req := Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/disbursement/v1_0/transfer/%s", referenceID),
		Headers: map[string]string{
			"Authorization":             "Bearer " + token,
			"X-Target-Environment":      s.config.TargetEnvironment,
			"Ocp-Apim-Subscription-Key": s.config.DisbursementKey,
		},
	}

	err = s.client.DoRequest(ctx, req, &result)
	if err != nil {
		return nil, fmt.Errorf("error checking transfer status: %w", err)
	}

	return &result, nil
}

// GetAccountBalance gets the balance of the account
func (s *DisbursementService) GetAccountBalance(ctx context.Context) (string, string, error) {
	// Get access token
	token, err := s.authService.GetAccessToken(ctx, "disbursement")
	if err != nil {
		return "", "", fmt.Errorf("error getting access token: %w", err)
	}

	var result struct {
		AvailableBalance string `json:"availableBalance"`
		Currency         string `json:"currency"`
	}

	req := Request{
		Method: http.MethodGet,
		Path:   "/disbursement/v1_0/account/balance",
		Headers: map[string]string{
			"Authorization":             "Bearer " + token,
			"X-Target-Environment":      s.config.TargetEnvironment,
			"Ocp-Apim-Subscription-Key": s.config.DisbursementKey,
		},
	}

	err = s.client.DoRequest(ctx, req, &result)
	if err != nil {
		return "", "", fmt.Errorf("error getting account balance: %w", err)
	}

	return result.AvailableBalance, result.Currency, nil
}

// GetAccountHolderInfo gets information about an account holder
func (s *DisbursementService) GetAccountHolderInfo(ctx context.Context, phone string) (*AccountHolderInfo, error) {
	// Format phone number if needed
	phone = formatPhoneNumber(phone)

	// Get access token
	token, err := s.authService.GetAccessToken(ctx, "disbursement")
	if err != nil {
		return nil, fmt.Errorf("error getting access token: %w", err)
	}

	var result AccountHolderInfo
	req := Request{
		Method: http.MethodGet,
		Path:   fmt.Sprintf("/disbursement/v1_0/accountholder/MSISDN/%s/basicuserinfo", phone),
		Headers: map[string]string{
			"Authorization":             "Bearer " + token,
			"X-Target-Environment":      s.config.TargetEnvironment,
			"Ocp-Apim-Subscription-Key": s.config.DisbursementKey,
		},
	}

	err = s.client.DoRequest(ctx, req, &result)
	if err != nil {
		return nil, fmt.Errorf("error getting account holder info: %w", err)
	}

	return &result, nil
}
