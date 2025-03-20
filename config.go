package gomomo

import (
	"fmt"
	"os"
)

// EnvironmentType represents the MTN MoMo environment (sandbox or production)
type EnvironmentType string

const (
	Sandbox    EnvironmentType = "sandbox"
	Production EnvironmentType = "production"
)

// Config holds the MTN MoMo API configuration
type Config struct {
	// Common configuration
	SubscriptionKey   string          // Primary subscription key for API access
	DisbursementKey   string          // Key for disbursement operations (can be same as SubscriptionKey)
	TargetEnvironment string          // Target environment (e.g., "sandbox", "prod", country code)
	CallbackHost      string          // Host for callback URLs
	APIUser           string          // API user ID (auto-generated in sandbox, provided in production)
	APIKey            string          // API key for the user
	Environment       EnvironmentType // Sandbox or Production
	Currency          string          // Default currency (EUR for sandbox, varies by country in production)

	// Environment-specific hosts
	Host string // API host URL
}

// NewConfig creates a new MTN MoMo configuration
func NewConfig(environment EnvironmentType, opts ...ConfigOption) (*Config, error) {
	// Default configuration based on environment
	config := &Config{
		Environment: environment,
		Currency:    "EUR", // Default for sandbox
	}

	// Set environment-specific defaults
	switch environment {
	case Sandbox:
		config.Host = "sandbox.momodeveloper.mtn.com"
		config.TargetEnvironment = "sandbox"
	case Production:
		config.Currency = "" // Will be determined by country in production
	}

	// Apply provided options
	for _, opt := range opts {
		opt(config)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// ConfigOption defines a function type for setting config options
type ConfigOption func(*Config)

// WithSubscriptionKey sets the subscription key
func WithSubscriptionKey(key string) ConfigOption {
	return func(c *Config) {
		c.SubscriptionKey = key
	}
}

// WithDisbursementKey sets the disbursement key
func WithDisbursementKey(key string) ConfigOption {
	return func(c *Config) {
		c.DisbursementKey = key
	}
}

// WithTargetEnvironment sets the target environment
func WithTargetEnvironment(env string) ConfigOption {
	return func(c *Config) {
		c.TargetEnvironment = env
	}
}

// WithCallbackHost sets the callback host
func WithCallbackHost(host string) ConfigOption {
	return func(c *Config) {
		c.CallbackHost = host
	}
}

// WithAPIUser sets the API user ID (usually for production)
func WithAPIUser(user string) ConfigOption {
	return func(c *Config) {
		c.APIUser = user
	}
}

// WithAPIKey sets the API key (usually for production)
func WithAPIKey(key string) ConfigOption {
	return func(c *Config) {
		c.APIKey = key
	}
}

// WithHost sets the API host URL
func WithHost(host string) ConfigOption {
	return func(c *Config) {
		c.Host = host
	}
}

// WithCurrency sets the default currency
func WithCurrency(currency string) ConfigOption {
	return func(c *Config) {
		c.Currency = currency
	}
}

// FromEnv loads configuration from environment variables
func FromEnv() ConfigOption {
	return func(c *Config) {
		if key := os.Getenv("MOMO_SUBSCRIPTION_KEY"); key != "" {
			c.SubscriptionKey = key
		}
		if key := os.Getenv("MOMO_DISBURSEMENT_KEY"); key != "" {
			c.DisbursementKey = key
		}
		if env := os.Getenv("MOMO_TARGET_ENVIRONMENT"); env != "" {
			c.TargetEnvironment = env
		}
		if host := os.Getenv("MOMO_CALLBACK_HOST"); host != "" {
			c.CallbackHost = host
		}
		if host := os.Getenv("MOMO_HOST"); host != "" {
			c.Host = host
		}
		if apiUser := os.Getenv("MOMO_API_USER"); apiUser != "" {
			c.APIUser = apiUser
		}
		if apiKey := os.Getenv("MOMO_API_KEY"); apiKey != "" {
			c.APIKey = apiKey
		}
		if currency := os.Getenv("MOMO_CURRENCY"); currency != "" {
			c.Currency = currency
		}
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.SubscriptionKey == "" {
		return fmt.Errorf("subscription key is required")
	}
	if c.TargetEnvironment == "" {
		return fmt.Errorf("target environment is required")
	}
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Environment == Production && c.APIUser == "" && c.APIKey == "" {
		return fmt.Errorf("API user and key are required for production")
	}
	if c.DisbursementKey == "" {
		// Use subscription key as default for disbursement if not specified
		c.DisbursementKey = c.SubscriptionKey
	}
	if c.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	return nil
}
