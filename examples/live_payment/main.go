// examples/live_payment/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sir-george2500/gomomo"
)

func main() {
	log.Println("Initializing MTN MoMo live mode test...")

	// Load production configuration
	config, err := gomomo.NewConfig(
		gomomo.Production,
		gomomo.WithSubscriptionKey(os.Getenv("MOMO_PROD_SUBSCRIPTION_KEY")),
		gomomo.WithDisbursementKey(os.Getenv("MOMO_PROD_DISBURSEMENT_KEY")),
		gomomo.WithCallbackHost(os.Getenv("MOMO_PROD_CALLBACK_HOST")),
		gomomo.WithHost(os.Getenv("MOMO_PROD_HOST")),
		gomomo.WithTargetEnvironment(os.Getenv("MOMO_PROD_TARGET_ENVIRONMENT")),
		gomomo.WithAPIUser(os.Getenv("MOMO_PROD_API_USER")),
		gomomo.WithAPIKey(os.Getenv("MOMO_PROD_API_KEY")),
		gomomo.WithCurrency(os.Getenv("MOMO_PROD_CURRENCY")),
	)
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	// Create MoMo client
	client := gomomo.NewMoMoClient(config)

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Test authentication first
	log.Println("Testing authentication...")
	token, err := client.Auth.GetAccessToken(ctx, "collection")
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}
	log.Printf("Successfully authenticated. Token: %s...", token[:10]+"...")

	// Get a valid phone number for testing
	var phone string
	log.Print("Enter a valid phone number to test with: ")
	_, err = fmt.Scanln(&phone)
	if err != nil {
		log.Fatalf("Error reading phone number: %v", err)
	}

	// Get the amount for testing
	var amount float64
	log.Print("Enter amount to test with (recommended small amount): ")
	_, err = fmt.Scanln(&amount)
	if err != nil {
		log.Fatalf("Error reading amount: %v", err)
	}

	// PART 1: Collection (Request to Pay)
	log.Printf("Testing collection with phone: %s, amount: %.2f %s",
		phone, amount, os.Getenv("MOMO_PROD_CURRENCY"))

	// Generate unique idempotency key
	idempotencyKey := gomomo.GenerateIdempotencyKey(
		"live_test",
		phone,
		fmt.Sprintf("%.2f", amount),
		time.Now().Format("20060102150405"),
	)
	log.Printf("Using idempotency key: %s", idempotencyKey)

	// Initiate payment request
	log.Println("Initiating payment request...")
	referenceID, err := client.Collection.RequestToPay(
		ctx,
		phone,
		amount,
		&gomomo.RequestToPayOptions{
			IdempotencyKey: idempotencyKey,
			PayerMessage:   "Test payment - please approve",
			PayeeNote:      "Thank you for testing our integration",
		},
	)
	if err != nil {
		log.Fatalf("Failed to initiate payment: %v", err)
	}
	log.Printf("Payment request initiated with reference ID: %s", referenceID)
	log.Println("CHECK YOUR PHONE TO APPROVE THE PAYMENT")

	// Poll for status with longer timeout and more attempts for production
	log.Println("Polling for transaction status...")
	maxPolls := 8
	pollInterval := 10 * time.Second

	var paymentSuccessful bool
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
		if status.Status == gomomo.Successful {
			log.Printf("Payment successful!")
			paymentSuccessful = true
			break
		} else if status.Status == gomomo.Failed || status.Status == gomomo.Rejected {
			log.Printf("Payment failed with status: %s", status.Status)
			break
		}

		time.Sleep(pollInterval)
	}

	// Only proceed with disbursement if collection was successful
	if paymentSuccessful {
		// PART 2: Disbursement (Transfer)
		log.Println("\nDo you want to test disbursement? (y/n)")
		var response string
		fmt.Scanln(&response)

		if response == "y" || response == "Y" {
			disbursementAmount := amount / 2 // Use half the amount for disbursement test
			log.Printf("Testing disbursement with amount: %.2f %s",
				disbursementAmount, os.Getenv("MOMO_PROD_CURRENCY"))

			disbursementIdempotencyKey := gomomo.GenerateIdempotencyKey(
				"live_disburse",
				phone,
				fmt.Sprintf("%.2f", disbursementAmount),
				time.Now().Format("20060102150405"),
			)

			transferReferenceID, err := client.Disbursement.Transfer(
				ctx,
				phone,
				disbursementAmount,
				&gomomo.TransferOptions{
					IdempotencyKey: disbursementIdempotencyKey,
					PayerMessage:   "Test disbursement",
					PayeeNote:      "Returning funds from test",
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
				if status.Status == gomomo.Successful {
					log.Printf("Transfer successful!")
					break
				} else if status.Status == gomomo.Failed ||
					status.Status == gomomo.Rejected {
					log.Printf("Transfer failed with status: %s", status.Status)
					break
				}

				time.Sleep(pollInterval)
			}
		}
	}

	log.Println("\nLive test completed!")
}
