package issuer

import "terraform-provider-boxer/pkg/generated/api/issuerClient"

// ProviderDataReader is an interface that defines methods to get and set the data needed for the issuer client.
type ProviderDataReader interface {
	// GetIssuerClient returns the issuer client.
	GetIssuerClient() *issuerClient.Client

	// GetHostName returns the host of the issuer.
	GetHostName() string
}
