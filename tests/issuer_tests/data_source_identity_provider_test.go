package issuer_tests

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	helpers "terraform-provider-boxer/tests"
	"testing"
)

func TestDataSourceIdentityProvider_reading(t *testing.T) {
	const resourceAddress = "data.boxer_identity_provider.example"
	const templateName = "data_source_identity_provider/data_source_identity_provider.tmpl.tf"

	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	services := helpers.NewLocalServices()
	token, err := helpers.GetExternalToken(services)
	if err != nil {
		t.Fatalf("failed to get external token: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: helpers.RenderTemplate(helpers.NewTestContext(randomName, services, token), templateName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(randomName),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("discovery_url"),
						knownvalue.StringExact(services.ExternalIdp.ClusterEndpoint),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("user_id_claim"),
						knownvalue.StringExact("preferred_username"),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("issuers"),
						knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact(services.ExternalIdp.Endpoint)}),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("audiences"),
						knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("account")}),
					),
				},
			},
		},
	})
}
