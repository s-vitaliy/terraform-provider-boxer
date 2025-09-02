package security

import (
	"context"
	"fmt"
	"io"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
)

// Validate that internalTokenReader implements InternalTokenReader.
var (
	_ InternalTokenReader = &internalTokenReader{}
)

// InternalTokenReader is an interface for reading internal tokens, which are used for authenticating API requests.
type InternalTokenReader interface {
	GetToken(ctx context.Context) (string, error)
}

// NewInternalTokenReader creates a new instance of internalTokenReader.
func NewInternalTokenReader(issuerClient *issuerClient.Client, identityProvider string) InternalTokenReader {
	return &internalTokenReader{
		internalTokenReader: issuerClient,
		identityProvider:    identityProvider,
	}
}

// internalTokenReader is an implementation of InternalTokenReader that uses an issuer client to obtain tokens.
type internalTokenReader struct {
	internalTokenReader *issuerClient.Client
	identityProvider    string
}

// GetToken retrieves an internal token using the configured issuer client and identity provider.
func (itr *internalTokenReader) GetToken(ctx context.Context) (string, error) {
	internalToken, err := itr.internalTokenReader.Token(ctx, issuerClient.TokenParams{IdentityProvider: itr.identityProvider})
	if err != nil {
		return "", fmt.Errorf("failed to get internal token: %w", err)
	}

	tokenValue, err := io.ReadAll(internalToken.Data)
	if err != nil {
		return "", fmt.Errorf("failed to read internal token data: %w", err)
	}
	return string(tokenValue), nil
}
