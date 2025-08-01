package security

import (
	"context"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
)

// Validate that emptySecuritySource implements issuer.SecuritySource.
var (
	_ issuerClient.SecuritySource = emptySecuritySource{}
)

// NewEmptySecuritySource creates a new instance of emptySecuritySource.
func NewEmptySecuritySource() issuerClient.SecuritySource {
	return emptySecuritySource{}
}

// emptySecuritySource is a no-op implementation of issuer.SecuritySource.
// It returns an empty token for any external security request.
type emptySecuritySource struct{}

// External implements the issuer.SecuritySource interface.
func (e emptySecuritySource) External(_ context.Context, _ issuerClient.OperationName) (issuerClient.External, error) {
	return issuerClient.External{Token: ""}, nil
}
