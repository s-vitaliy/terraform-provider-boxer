package issuer_tests

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"terraform-provider-boxer/internal/tests"
	"terraform-provider-boxer/internal/tests/assertions"
	"testing"
)

func TestDataSourceCedarSchema_reading(t *testing.T) {
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
				Config: tests.RenderTemplate(tests.NewTestContext(randomName, services, token), "data_source_cedar_schema.tmpl"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.boxer_issuer_cedar_schema.example",
						tfjsonpath.New("id"),
						knownvalue.StringExact(randomName),
					),
					assertions.ValidateSchemaIsParseable("data.boxer_issuer_cedar_schema.example"),
				},
			},
		},
	})
}
