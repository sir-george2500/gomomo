package gomomo

// MoMoClient is the main client for interacting with MTN MoMo API
type MoMoClient struct {
	Config       *Config
	Auth         *AuthService
	Collection   *CollectionService
	Disbursement *DisbursementService
}

// NewMoMoClient creates a new MTN MoMo client
func NewMoMoClient(config *Config) *MoMoClient {
	client := NewClient(config)
	authService := NewAuthService(client, config)

	return &MoMoClient{
		Config:       config,
		Auth:         authService,
		Collection:   NewCollectionService(client, config, authService),
		Disbursement: NewDisbursementService(client, config, authService),
	}
}

// InitFromEnv creates a new MoMoClient from environment variables
func InitFromEnv(environment EnvironmentType) (*MoMoClient, error) {
	config, err := NewConfig(environment, FromEnv())
	if err != nil {
		return nil, err
	}

	return NewMoMoClient(config), nil
}
