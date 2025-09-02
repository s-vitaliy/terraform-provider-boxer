package security

import (
	"context"
	"fmt"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"
)

// Validate that validatorInternalSecuritySource implements issuer.SecuritySource.
var (
	_ validatorClient.SecuritySource = validatorInternalSecuritySource{}
)

// NewValidatorInternalSecuritySource creates a new instance of validatorInternalSecuritySource.
// internalTokenReader is used to obtain internal tokens from the issuer.
// identityProvider is the ID of the identity provider to use for obtaining tokens.
func NewValidatorInternalSecuritySource(internalTokenReader InternalTokenReader) validatorClient.SecuritySource {
	return validatorInternalSecuritySource{
		internalTokenReader: internalTokenReader,
	}
}

// validatorInternalSecuritySource is a no-op implementation of issuer.SecuritySource.
// It returns an empty token for any external security request.
type validatorInternalSecuritySource struct {
	internalTokenReader InternalTokenReader
}

func (e validatorInternalSecuritySource) Internal(ctx context.Context, _ validatorClient.OperationName) (validatorClient.Internal, error) {
	token, err := e.internalTokenReader.GetToken(ctx)
	if err != nil {
		return validatorClient.Internal{}, fmt.Errorf("failed to get internal token: %w", err)
	}
	return validatorClient.Internal{Token: token}, nil
}
