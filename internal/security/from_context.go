package security

import (
	"context"
	"fmt"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
)

type ContextKey string

// Validate that securitySourceFromContext implements issuer.SecuritySource.
var (
	_ issuerClient.SecuritySource = securitySourceFromContext{}
)

// NewSecuritySourceFromContext creates a new instance of securitySourceFromContext.
func NewSecuritySourceFromContext(contextField ContextKey) issuerClient.SecuritySource {
	return securitySourceFromContext{
		field: contextField,
	}
}

// securitySourceFromContext is a no-op implementation of issuer.SecuritySource.
// It returns an empty token for any external security request.
type securitySourceFromContext struct {
	field ContextKey
}

// External implements the issuer.SecuritySource interface.
func (e securitySourceFromContext) External(ctx context.Context, _ issuerClient.OperationName) (issuerClient.External, error) {
	token, ok := ctx.Value(e.field).(string)
	if !ok {
		return issuerClient.External{}, fmt.Errorf("context field %q not found or not a string", e.field)
	}
	return issuerClient.External{Token: token}, nil
}
