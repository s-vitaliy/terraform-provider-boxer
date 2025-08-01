package security

import (
	"context"
	"fmt"
	issuer "terraform-provider-boxer/pkg/generated/api"
)

type ContextKey string

// Validate that securitySourceFromContext implements issuer.SecuritySource.
var (
	_ issuer.SecuritySource = securitySourceFromContext{}
)

// NewSecuritySourceFromContext creates a new instance of securitySourceFromContext.
func NewSecuritySourceFromContext(contextField ContextKey) issuer.SecuritySource {
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
func (e securitySourceFromContext) External(ctx context.Context, _ issuer.OperationName) (issuer.External, error) {
	token, ok := ctx.Value(e.field).(string)
	if !ok {
		return issuer.External{}, fmt.Errorf("context field %q not found or not a string", e.field)
	}
	return issuer.External{Token: token}, nil
}
