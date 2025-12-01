package issuer

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
)

func getResourceIssuerClient(request resource.ConfigureRequest, response *resource.ConfigureResponse) *issuerClient.Client {
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
	if data.GetIssuerClient() == nil { // coverage-ignore
		response.Diagnostics.AddError(
			"Invalid Issuer Client",
			"The issuer client must not be nil. This is most likely the bug in the provider implementation.",
		)
		return nil
	}
	client := data.GetIssuerClient()
	return client
}
