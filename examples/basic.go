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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test API user creation and authentication
	log.Println("Testing authentication...")
	apiUser, err := client.Auth.CreateAPIUser(ctx)
	if err != nil {
		log.Fatalf("Failed to create API user: %v", err)
	}
	log.Printf("Created API user: %s", apiUser)

	apiKey, err := client.Auth.CreateAPIKey(ctx, apiUser)
	if err != nil {
		log.Fatalf("Failed to create API key: %v", err)
	}
	log.Printf("Created API key: %s", apiKey)

	token, err := client.Auth.GetAccessToken(ctx, "collection")
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}
	log.Printf("Got access token: %s", token[:10]+"...")

	fmt.Println("\nAuthentication successful!")
	fmt.Println("=========================")

	// The actual payment integration would go here, but we'll skip it
	// for now since it requires a phone number that can receive the payment prompt
	fmt.Println("\nBasic API connectivity test passed!")
}
