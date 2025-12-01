package issuer_tests

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	helpers "terraform-provider-boxer/tests"
	"testing"
)

func TestResourceBoxerExternalIdentity_creation(t *testing.T) {
	const resourceAddress = "boxer_external_identity.example"
	const templateName = "resource_boxer_external_identity/resource_boxer_external_identity.tmpl.tf"

	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	services := helpers.NewLocalServices()
	token, err := helpers.GetExternalToken(services)
	if err != nil {
		t.Fatalf("failed to get external token: %s", err)
	}
	testContext := helpers.NewTestContext(randomName, services, token)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: helpers.RenderTemplate(testContext, templateName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("id"),
						knownvalue.StringExact(randomName),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("identity_provider"),
						knownvalue.StringExact(randomName),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("principal").AtMapKey("principal_id"),
						knownvalue.StringExact("PhotoApp::User::\"alice\""),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("principal").AtMapKey("schema_id"),
						knownvalue.StringExact(fmt.Sprintf("%s-issuer", randomName)),
					),
					statecheck.ExpectKnownValue(
						resourceAddress,
						tfjsonpath.New("validator_schema_id"),
						knownvalue.StringExact(fmt.Sprintf("%s-validator", randomName)),
					),
				},
			},
		},
	})
}
