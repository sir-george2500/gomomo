package gomomo

import (
	"strings"
	"time"
)

// PartyIDType represents the type of party ID
type PartyIDType string

const (
	// MSISDN is a mobile number
	MSISDN PartyIDType = "MSISDN"
	// Email is an email address
	Email PartyIDType = "EMAIL"
	// Party is a party code
	Party PartyIDType = "PARTY_CODE"
)

// PartyInfo represents a payer or payee in a transaction
type PartyInfo struct {
	PartyIDType PartyIDType `json:"partyIdType"`
	PartyID     string      `json:"partyId"`
}

// RequestToPayPayload represents the payload for a collection request
type RequestToPayPayload struct {
	Amount       string    `json:"amount"`
	Currency     string    `json:"currency"`
	ExternalID   string    `json:"externalId"`
	Payer        PartyInfo `json:"payer"`
	PayerMessage string    `json:"payerMessage"`
	PayeeNote    string    `json:"payeeNote"`
}

// TransferPayload represents the payload for a disbursement request
type TransferPayload struct {
	Amount       string    `json:"amount"`
	Currency     string    `json:"currency"`
	ExternalID   string    `json:"externalId"`
	Payee        PartyInfo `json:"payee"`
	PayerMessage string    `json:"payerMessage"`
	PayeeNote    string    `json:"payeeNote"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	// Pending means the transaction is being processed
	Pending TransactionStatus = "PENDING"
	// Successful means the transaction was successful
	Successful TransactionStatus = "SUCCESSFUL"
	// Failed means the transaction failed
	Failed TransactionStatus = "FAILED"
	// Rejected means the transaction was rejected
	Rejected TransactionStatus = "REJECTED"
	// Timeout means the transaction timed out
	Timeout TransactionStatus = "TIMEOUT"
)

// TransactionStatusResponse represents the response from a transaction status check
type TransactionStatusResponse struct {
	Amount                 string            `json:"amount"`
	Currency               string            `json:"currency"`
	ExternalID             string            `json:"externalId"`
	Payer                  PartyInfo         `json:"payer,omitempty"`
	Payee                  PartyInfo         `json:"payee,omitempty"`
	PayerMessage           string            `json:"payerMessage,omitempty"`
	PayeeNote              string            `json:"payeeNote,omitempty"`
	Status                 TransactionStatus `json:"status"`
	Reason                 string            `json:"reason,omitempty"`
	FinancialTransactionID string            `json:"financialTransactionId,omitempty"`
}

// AccountHolderInfo represents basic information about an account holder
type AccountHolderInfo struct {
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Birthdate  string `json:"birthdate"`
	Locale     string `json:"locale"`
	Gender     string `json:"gender"`
	Status     string `json:"status"`
}

// GenerateIdempotencyKey creates a unique idempotency key
// The format can be customized based on your needs
func GenerateIdempotencyKey(prefix string, uniqueElements ...string) string {
	elements := append([]string{prefix}, uniqueElements...)
	elements = append(elements, time.Now().Format("20060102150405"))
	return strings.Join(elements, "_")
}
