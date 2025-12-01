package validator

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"
)

func getDataSourceValidatorClient(request datasource.ConfigureRequest, response *datasource.ConfigureResponse) *validatorClient.Client {
	if request.ProviderData == nil {
		return nil
	}
	data, ok := request.ProviderData.(ProviderDataReader)
	if !ok { // coverage-ignore
		response.Diagnostics.AddError(
			"Invalid Provider Data",
			"The provider data must be of type *BoxerProviderData, but was %s. This is most likely the bug in the provider implementation.",
		)
		return nil
	}
	if data.GetValidatorClient() == nil { // coverage-ignore
		response.Diagnostics.AddError(
			"Invalid Issuer Client",
			"The issuer client must not be nil. This is most likely the bug in the provider implementation.",
		)
		return nil
	}
	client := data.GetValidatorClient()
	return client
}
