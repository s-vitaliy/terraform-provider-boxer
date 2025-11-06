package tests

// ExternalIdentityProviderCredentials holds credentials for the external identity provider used in tests.
type ExternalIdentityProviderCredentials struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	GrantType    string
}

// ExternalIdentityProviderEndpoints holds configuration for the external identity provider used in tests.
type ExternalIdentityProviderEndpoints struct {
	Endpoint        string
	ClusterEndpoint string
	Credentials     ExternalIdentityProviderCredentials
}

// Services holds the endpoints for the mock services used in tests.
type Services struct {
	// Endpoints for the mock services
	issuerEndpoint    string
	validatorEndpoint string
	ExternalIdp       ExternalIdentityProviderEndpoints
}

// NewLocalServices returns a Services struct configured to point to local mock services.
// This is intended for use in tests only.
// Note: the endpoints must match those configured in the local mock services (see integration_tests folder).
func NewLocalServices() *Services {
	return &Services{
		issuerEndpoint:    "http://localhost:5555/issuer",
		validatorEndpoint: "http://localhost:5555/validator",
		ExternalIdp: ExternalIdentityProviderEndpoints{
			Endpoint:        "http://localhost:5555/auth",
			ClusterEndpoint: "http://keycloak-keycloakx-http/auth/realms/master/",
			Credentials: ExternalIdentityProviderCredentials{
				ClientID:     "test_client",
				ClientSecret: "test_client_secret",
				Username:     "test_root",
				Password:     "test-root-password",
				GrantType:    "password",
			},
		},
	}
}
