package security

import (
	"context"
	"fmt"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
)

type ContextKey string

// Validate that issuerContextSecuritySource implements issuer.SecuritySource.
var (
	_ issuerClient.SecuritySource = issuerContextSecuritySource{}
)

// NewIssuerContextSecuritySource creates a new instance of issuerContextSecuritySource.
func NewIssuerContextSecuritySource(contextField ContextKey) issuerClient.SecuritySource {
	return issuerContextSecuritySource{
		field: contextField,
	}
}

// issuerContextSecuritySource is a no-op implementation of issuer.SecuritySource.
// It returns an empty token for any external security request.
type issuerContextSecuritySource struct {
	field ContextKey
}

func (e issuerContextSecuritySource) Internal(ctx context.Context, _ issuerClient.OperationName) (issuerClient.Internal, error) {
	token, ok := ctx.Value(e.field).(string)
	if !ok {
		return issuerClient.Internal{}, fmt.Errorf("context field %q not found or not a string", e.field)
	}
	return issuerClient.Internal{Token: token}, nil
}

// External implements the issuer.SecuritySource interface.
func (e issuerContextSecuritySource) External(ctx context.Context, _ issuerClient.OperationName) (issuerClient.External, error) {
	token, ok := ctx.Value(e.field).(string)
	if !ok {
		return issuerClient.External{}, fmt.Errorf("context field %q not found or not a string", e.field)
	}
	return issuerClient.External{Token: token}, nil
}
