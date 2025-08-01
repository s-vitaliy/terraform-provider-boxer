package internal

import (
	"terraform-provider-boxer/internal/issuer"
	"terraform-provider-boxer/internal/validator"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"
)

var (
	_ issuer.ProviderDataReader    = &BoxerProviderData{}
	_ validator.ProviderDataReader = &BoxerProviderData{}
)

type BoxerProviderData struct {
	issuerHost      string
	issuerClient    *issuerClient.Client
	validatorClient *validatorClient.Client
}

func (b *BoxerProviderData) GetIssuerClient() *issuerClient.Client {
	return b.issuerClient
}

func (b *BoxerProviderData) GetHostName() string {
	return b.issuerHost
}

func (b *BoxerProviderData) GetValidatorClient() *validatorClient.Client {
	return b.validatorClient
}
