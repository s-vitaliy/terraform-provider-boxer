package issuer_tests

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"terraform-provider-boxer/internal/tests"
	"testing"
)

func TestDataSourceIdentityProvider_reading(t *testing.T) {
	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	services := tests.NewLocalServices()
	token, err := tests.GetExternalToken(services)
	if err != nil {
		t.Fatalf("failed to get external token: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tests.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tests.RenderTemplate(tests.NewTestContext(randomName, services, token), "data_source_identity_provider.tmpl"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.boxer_identity_provider.example",
						tfjsonpath.New("id"),
						knownvalue.StringExact(randomName),
					),
					statecheck.ExpectKnownValue(
						"data.boxer_identity_provider.example",
						tfjsonpath.New("discovery_url"),
						knownvalue.StringExact(services.ExternalIdp.ClusterEndpoint),
					),
					statecheck.ExpectKnownValue(
						"data.boxer_identity_provider.example",
						tfjsonpath.New("user_id_claim"),
						knownvalue.StringExact("preferred_username"),
					),
					statecheck.ExpectKnownValue(
						"data.boxer_identity_provider.example",
						tfjsonpath.New("issuers"),
						knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact(services.ExternalIdp.Endpoint)}),
					),
					statecheck.ExpectKnownValue(
						"data.boxer_identity_provider.example",
						tfjsonpath.New("audiences"),
						knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("account")}),
					),
				},
			},
		},
	})
}
