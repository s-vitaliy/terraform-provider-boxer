package validator

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"reflect"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"
)

func getResourceValidatorClient(request resource.ConfigureRequest, response *resource.ConfigureResponse) *validatorClient.Client {
	if request.ProviderData == nil {
		return nil
	}
	data, ok := request.ProviderData.(ProviderDataReader)
	if !ok { // coverage-ignore
		response.Diagnostics.AddError(
			"Invalid Provider Data",
			"The provider data must be of type ProviderDataReader,"+
				fmt.Sprintf("but was %s. ", reflect.TypeOf(request.ProviderData).String())+
				"This is most likely the bug in the provider implementation.",
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
