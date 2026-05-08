package issuer_tests

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"regexp"
	helpers "terraform-provider-boxer/tests"
	"testing"
)

func TestResourceInvalidCedarSchema_creation(t *testing.T) {
	const templateName = "resource_invalid_cedar_schema/resource_invalid_cedar_schema.tmpl.tf"

	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	services := helpers.NewLocalServices()
	token, err := helpers.GetExternalToken(services)
	testContext := helpers.NewTestContext(randomName, services, token)
	if err != nil {
		t.Fatalf("failed to get external token: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      helpers.RenderTemplate(testContext, templateName),
				ExpectError: regexp.MustCompile(`Invalid Cedar schema.`),
			},
		},
	})
}
