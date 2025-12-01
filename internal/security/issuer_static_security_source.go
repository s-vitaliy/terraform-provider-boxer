package security

import (
	"context"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
)

// Validate that issuerStaticSecuritySource implements issuer.SecuritySource.
var (
	_ issuerClient.SecuritySource = issuerStaticSecuritySource{}
)

// IssuerStaticSecuritySource creates a new instance of issuerStaticSecuritySource.
func IssuerStaticSecuritySource(value string) issuerClient.SecuritySource {
	return issuerStaticSecuritySource{
		value: value,
	}
}

// issuerStaticSecuritySource is a no-op implementation of issuer.SecuritySource.
// It returns an empty token for any external security request.
type issuerStaticSecuritySource struct {
	value string
}

func (e issuerStaticSecuritySource) Internal(_ context.Context, _ issuerClient.OperationName) (issuerClient.Internal, error) { // coverage-ignore
	return issuerClient.Internal{Token: e.value}, nil
}

// External implements the issuer.SecuritySource interface.
func (e issuerStaticSecuritySource) External(_ context.Context, _ issuerClient.OperationName) (issuerClient.External, error) {
	return issuerClient.External{Token: e.value}, nil
}
