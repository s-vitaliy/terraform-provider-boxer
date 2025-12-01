package security

import (
	"context"
	"fmt"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
)

// Validate that issuerInternalSecuritySource implements issuer.SecuritySource.
var (
	_ issuerClient.SecuritySource = issuerInternalSecuritySource{}
)

// NewIssuerInternalSecuritySource creates a new instance of issuerInternalSecuritySource.
// internalTokenReader is used to obtain internal tokens from the issuer.
// identityProvider is the ID of the identity provider to use for obtaining tokens.
func NewIssuerInternalSecuritySource(internalTokenReader InternalTokenReader) issuerClient.SecuritySource {
	return issuerInternalSecuritySource{
		internalTokenReader: internalTokenReader,
	}
}

// issuerInternalSecuritySource is a no-op implementation of issuer.SecuritySource.
// It returns an empty token for any external security request.
type issuerInternalSecuritySource struct {
	internalTokenReader InternalTokenReader
}

func (e issuerInternalSecuritySource) Internal(ctx context.Context, _ issuerClient.OperationName) (issuerClient.Internal, error) {
	token, err := e.internalTokenReader.GetToken(ctx)
	if err != nil { // coverage-ignore
		return issuerClient.Internal{}, fmt.Errorf("failed to get internal token: %w", err)
	}
	return issuerClient.Internal{Token: token}, nil
}

// External implements the issuer.SecuritySource interface.
func (e issuerInternalSecuritySource) External(_ context.Context, _ issuerClient.OperationName) (issuerClient.External, error) {
	return issuerClient.External{}, fmt.Errorf("external security not supported in this context")
}
