package security

import (
	"context"
	issuer "terraform-provider-boxer/pkg/generated/api"
)

// Validate that emptySecuritySource implements issuer.SecuritySource.
var (
	_ issuer.SecuritySource = emptySecuritySource{}
)

// NewEmptySecuritySource creates a new instance of emptySecuritySource.
func NewEmptySecuritySource() issuer.SecuritySource {
	return emptySecuritySource{}
}

// emptySecuritySource is a no-op implementation of issuer.SecuritySource.
// It returns an empty token for any external security request.
type emptySecuritySource struct{}

// External implements the issuer.SecuritySource interface.
func (e emptySecuritySource) External(_ context.Context, _ issuer.OperationName) (issuer.External, error) {
	return issuer.External{Token: ""}, nil
}
