# MTN Mobile Money Go Package

A simple, clean, and easy-to-use Go package for integrating with MTN Mobile Money API for both collections (receiving payments) and disbursements (sending money).

## Features

- Support for both Sandbox and Production environments
- Easy configuration with environment variables or code
- Collection API (Request to Pay)
- Disbursement API (Transfer)
- Account balance and user information
- Transaction status checking
- Automatic token management
- Idempotency support to prevent duplicate transactions
- Comprehensive error handling

## Installation

```bash
go get github.com/sir-george2500/gomomo
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	gomomo "github.com/sir-george2500/gomomo"
)

func main() {
	// Initialize from environment variables
	client, err := gomomo.InitFromEnv(gomomo.Sandbox)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Request a payment
	referenceID, err := client.Collection.RequestToPay(
		context.Background(),
		"231123456789",
		10.00,
		&gomomo.RequestToPayOptions{
			PayerMessage: "Payment for goods",
			PayeeNote: "Thank you",
		},
	)
	if err != nil {
		log.Fatalf("Failed to request payment: %v", err)
	}

	fmt.Printf("Payment requested with reference ID: %s\n", referenceID)
}
```

## Configuration

### Environment Variables

You can configure the package using environment variables:

```
# Sandbox Environment
MOMO_SUBSCRIPTION_KEY=your-sandbox-subscription-key
MOMO_DISBURSEMENT_KEY=your-sandbox-disbursement-key
MOMO_TARGET_ENVIRONMENT=sandbox
MOMO_CALLBACK_HOST=https://your-callback-host.com
MOMO_HOST=sandbox.momodeveloper.mtn.com
MOMO_CURRENCY=EUR

# Production Environment
MOMO_PROD_SUBSCRIPTION_KEY=your-production-subscription-key
MOMO_PROD_DISBURSEMENT_KEY=your-production-disbursement-key
MOMO_PROD_CALLBACK_HOST=your-production-callback-host
MOMO_PROD_HOST=your-production-host
MOMO_PROD_TARGET_ENVIRONMENT=your-production-target-environment
MOMO_PROD_API_USER=your-production-api-user
MOMO_PROD_API_KEY=your-production-api-key
MOMO_PROD_CURRENCY=your-production-currency
```

### Code Configuration

Or configure directly in code:

```go
// Sandbox configuration
config, err := gomomo.NewConfig(
    gomomo.Sandbox,
    gomomo.WithSubscriptionKey("your-subscription-key"),
    gomomo.WithDisbursementKey("your-disbursement-key"),
    gomomo.WithCallbackHost("https://your-callback-host.com"),
    gomomo.WithHost("sandbox.momodeveloper.mtn.com"),
    gomomo.WithTargetEnvironment("sandbox"),
    gomomo.WithCurrency("EUR"),
)

// Production configuration
config, err := gomomo.NewConfig(
    gomomo.Production,
    gomomo.WithSubscriptionKey("your-production-subscription-key"),
    gomomo.WithDisbursementKey("your-production-disbursement-key"),
    gomomo.WithCallbackHost("your-production-callback-host"),
    gomomo.WithHost("your-production-host"),
    gomomo.WithTargetEnvironment("your-production-target-environment"),
    gomomo.WithAPIUser("your-production-api-user"),
    gomomo.WithAPIKey("your-production-api-key"),
    gomomo.WithCurrency("your-production-currency"),
)
```

## Usage Examples

### Collection Service (Receiving Payments)

```go
// Request a payment with idempotency key
referenceID, err := client.Collection.RequestToPay(
    ctx,
    phone,
    amount,
    &gomomo.RequestToPayOptions{
        IdempotencyKey: "unique-payment-id-12345",
        PayerMessage: "Payment for order #12345",
        PayeeNote: "Thank you for your order",
    },
)

// Check transaction status
status, err := client.Collection.GetTransactionStatus(ctx, referenceID)
fmt.Printf("Transaction status: %s\n", status.Status)

// Get account balance
balance, currency, err := client.Collection.GetAccountBalance(ctx)
fmt.Printf("Balance: %s %s\n", balance, currency)

// Get account holder information
accountInfo, err := client.Collection.GetAccountHolderInfo(ctx, phone)
fmt.Printf("Account holder: %s %s\n", accountInfo.GivenName, accountInfo.FamilyName)
```

### Disbursement Service (Sending Money)

```go
// Send money with idempotency key
referenceID, err := client.Disbursement.Transfer(
    ctx,
    phone,
    amount,
    &gomomo.TransferOptions{
        IdempotencyKey: "unique-transfer-id-12345",
        PayerMessage: "Refund for order #12345",
        PayeeNote: "Refund processed",
    },
)

// Check transfer status
status, err := client.Disbursement.GetTransferStatus(ctx, referenceID)
fmt.Printf("Transfer status: %s\n", status.Status)

// Get account balance
balance, currency, err := client.Disbursement.GetAccountBalance(ctx)
fmt.Printf("Disbursement balance: %s %s\n", balance, currency)

// Get account holder information
accountInfo, err := client.Disbursement.GetAccountHolderInfo(ctx, phone)
fmt.Printf("Account holder: %s %s\n", accountInfo.GivenName, accountInfo.FamilyName)
```

## Idempotency Support

The package includes built-in support for idempotency to prevent duplicate transactions:

```go
// Generate an idempotency key based on business logic
idempotencyKey := gomomo.GenerateIdempotencyKey(
    "payment",
    userID,
    orderID,
    time.Now().Format("20060102"),
)

// Use the key in your request
referenceID, err := client.Collection.RequestToPay(
    ctx,
    phone,
    amount,
    &gomomo.RequestToPayOptions{
        IdempotencyKey: idempotencyKey,
        // Other options...
    },
)
```

## Troubleshooting

### IP Whitelisting for Disbursement

For production disbursement operations, MTN MoMo requires your IP address to be whitelisted. If you encounter a 403 Forbidden error with a message like "IP not authorized to utilize Disbursement API", you'll need to:

1. Identify your public IP address
2. Contact your MTN Account Manager with your public IP address
3. Wait for confirmation (typically 1-2 business days)
4. Test again after whitelisting is complete

Collection operations typically don't require IP whitelisting and should work without this step.

### Common Errors

- **401 Unauthorized**: Check your subscription keys and API credentials
- **403 Forbidden**: Check IP whitelisting for disbursement operations
- **404 Not Found**: Verify the API endpoint and reference IDs
- **500 Internal Server Error**: Contact MTN support

## Examples

The package includes several examples in the `examples` directory:

- `basic.go`: Basic authentication test
- `payment/main.go`: Sandbox payment flow
- `live_payment/main.go`: Production payment flow

To run the examples:

```bash
# Set environment variables
source examples/.env

# Run basic authentication test
cd examples
go run basic.go

# Run sandbox payment flow
cd examples/payment
go run main.go

# Run production payment flow (after setting production environment variables)
cd examples/live_payment
go run main.go
```

## Getting MTN MoMo API Credentials

1. Register for a developer account at [MTN MoMo Developer Portal](https://momodeveloper.mtn.com/)
2. Subscribe to Collection and/or Disbursement products
3. Get your subscription keys and other credentials
4. For production, contact MTN to get production credentials

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request anytime.
