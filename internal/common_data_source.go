package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	issuer "terraform-provider-boxer/pkg/generated/api"
)

func getDataSourceIssuerClient(request datasource.ConfigureRequest, response *datasource.ConfigureResponse) *issuer.Client {
	if request.ProviderData == nil {
		return nil
	}
	data, ok := request.ProviderData.(*BoxerProviderData)
	if !ok {
		response.Diagnostics.AddError(
			"Invalid Provider Data",
			"The provider data must be of type *BoxerProviderData, but was %s. This is most likely the bug in the provider implementation.",
		)
		return nil
	}
	if data.issuerClient == nil {
		response.Diagnostics.AddError(
			"Invalid Issuer Client",
			"The issuer client must not be nil. This is most likely the bug in the provider implementation.",
		)
		return nil
	}
	client := data.issuerClient
	return client
}
