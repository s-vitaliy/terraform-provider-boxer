package issuer

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
)

func getDataSourceIssuerClient(request datasource.ConfigureRequest, response *datasource.ConfigureResponse) *issuerClient.Client {
	if request.ProviderData == nil {
		return nil
	}
	data, ok := request.ProviderData.(ProviderDataReader)
	if !ok {
		response.Diagnostics.AddError(
			"Invalid Provider Data",
			"The provider data must be of type *BoxerProviderData, but was %s. This is most likely the bug in the provider implementation.",
		)
		return nil
	}
	if data.GetIssuerClient() == nil {
		response.Diagnostics.AddError(
			"Invalid Issuer Client",
			"The issuer client must not be nil. This is most likely the bug in the provider implementation.",
		)
		return nil
	}
	client := data.GetIssuerClient()
	return client
}

func getDataSourceIssuerHost(request datasource.ConfigureRequest, response *datasource.ConfigureResponse) string {
	if request.ProviderData == nil {
		return ""
	}
	data, ok := request.ProviderData.(ProviderDataReader)
	if !ok {
		response.Diagnostics.AddError(
			"Invalid Provider Data",
			"The provider data must be of type *BoxerProviderData, but was %s. This is most likely the bug in the provider implementation.",
		)
		return ""
	}
	if data.GetHostName() == "" {
		response.Diagnostics.AddError(
			"Invalid Issuer Host",
			"The issuer host must not be empty. This is most likely the bug in the provider implementation.",
		)
		return ""
	}
	return data.GetHostName()
}
