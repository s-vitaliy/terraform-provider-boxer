package validator

import "terraform-provider-boxer/pkg/generated/api/validatorClient"

// ProviderDataReader is an interface that defines methods to get and set the data needed for the issuer client.
type ProviderDataReader interface {
	// GetValidatorClient returns the issuer client.
	GetValidatorClient() *validatorClient.Client
}
