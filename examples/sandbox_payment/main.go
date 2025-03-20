// examples/payment
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/sir-george2500/gomomo"
)

func main() {
	// Load configuration from environment variables
	config, err := gomomo.NewConfig(
		gomomo.Sandbox,
		gomomo.WithSubscriptionKey(os.Getenv("MOMO_SUBSCRIPTION_KEY")),
		gomomo.WithDisbursementKey(os.Getenv("MOMO_DISBURSEMENT_KEY")),
		gomomo.WithCallbackHost(os.Getenv("MOMO_CALLBACK_HOST")),
		gomomo.WithHost(os.Getenv("MOMO_HOST")),
		gomomo.WithTargetEnvironment(os.Getenv("MOMO_TARGET_ENVIRONMENT")),
		gomomo.WithCurrency("EUR"),
	)
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	// Create MoMo client
	client := gomomo.NewMoMoClient(config)

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Example phone number for sandbox testing
	// In sandbox, use a number that will auto-respond to your request
	phone := "46733123454" // This is an example - check MTN docs for valid test numbers
	amount := 5.00

	// Create idempotency key for the request
	idempotencyKey := gomomo.GenerateIdempotencyKey("test_payment", time.Now().Format("20060102150405"))
	log.Printf("Using idempotency key: %s", idempotencyKey)

	// Initiate payment request
	log.Println("Initiating payment request...")
	referenceID, err := client.Collection.RequestToPay(
		ctx,
		phone,
		amount,
		&gomomo.RequestToPayOptions{
			IdempotencyKey: idempotencyKey,
			PayerMessage:   "Test payment",
			PayeeNote:      "Thank you for testing",
		},
	)
	if err != nil {
		log.Fatalf("Failed to initiate payment: %v", err)
	}
	log.Printf("Payment request initiated with reference ID: %s", referenceID)

	// Poll for status a few times
	log.Println("Polling for transaction status...")
	maxPolls := 6
	pollInterval := 5 * time.Second

	for i := 0; i < maxPolls; i++ {
		log.Printf("Polling attempt %d/%d...", i+1, maxPolls)
		status, err := client.Collection.GetTransactionStatus(ctx, referenceID)
		if err != nil {
			log.Printf("Error checking status: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		log.Printf("Transaction status: %s", status.Status)

		// If we have a final status, break the loop
		if status.Status == gomomo.Successful ||
			status.Status == gomomo.Failed ||
			status.Status == gomomo.Rejected {
			log.Printf("Final status reached: %s", status.Status)
			break
		}

		time.Sleep(pollInterval)
	}

	// Try a disbursement operation (transfer)
	log.Println("\nTesting disbursement...")

	disbursementIdempotencyKey := gomomo.GenerateIdempotencyKey("test_disbursement", time.Now().Format("20060102150405"))
	transferReferenceID, err := client.Disbursement.Transfer(
		ctx,
		phone,
		2.50,
		&gomomo.TransferOptions{
			IdempotencyKey: disbursementIdempotencyKey,
			PayerMessage:   "Test disbursement",
			PayeeNote:      "Funds received - test",
		},
	)
	if err != nil {
		log.Fatalf("Failed to initiate transfer: %v", err)
	}
	log.Printf("Transfer initiated with reference ID: %s", transferReferenceID)

	// Poll for transfer status
	log.Println("Polling for transfer status...")
	for i := 0; i < maxPolls; i++ {
		log.Printf("Polling attempt %d/%d...", i+1, maxPolls)
		status, err := client.Disbursement.GetTransferStatus(ctx, transferReferenceID)
		if err != nil {
			log.Printf("Error checking transfer status: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		log.Printf("Transfer status: %s", status.Status)

		// If we have a final status, break the loop
		if status.Status == gomomo.Successful ||
			status.Status == gomomo.Failed ||
			status.Status == gomomo.Rejected {
			log.Printf("Final transfer status reached: %s", status.Status)
			break
		}

		time.Sleep(pollInterval)
	}

	log.Println("\nPayment flow test completed!")
}
