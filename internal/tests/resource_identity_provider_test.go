package tests

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"html/template"
	"testing"
)

func TestIdentityProvider_creation(t *testing.T) {
	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	services := NewLocalServices()
	token, err := getExternalToken(services)
	if err != nil {
		t.Fatalf("failed to get external token: %s", err)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleResource(NewTestContext(randomName, services, token)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"boxer_identity_provider.example",
						tfjsonpath.New("id"),
						knownvalue.StringExact(randomName),
					),
					statecheck.ExpectKnownValue(
						"boxer_identity_provider.example",
						tfjsonpath.New("name"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"boxer_identity_provider.example",
						tfjsonpath.New("discovery_url"),
						knownvalue.StringExact(services.ExternalIdp.ClusterEndpoint),
					),
					statecheck.ExpectKnownValue(
						"boxer_identity_provider.example",
						tfjsonpath.New("user_id_claim"),
						knownvalue.StringExact("preferred_username"),
					),
					statecheck.ExpectKnownValue(
						"boxer_identity_provider.example",
						tfjsonpath.New("issuers"),
						knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact(services.ExternalIdp.Endpoint)}),
					),
					statecheck.ExpectKnownValue(
						"boxer_identity_provider.example",
						tfjsonpath.New("audiences"),
						knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("account")}),
					),
				},
			},
		},
	})
}

func testAccExampleResource(testContext *TestContext) string {
	tpl, err := template.New("configuration").ParseFiles("templates/identity_provider.tmpl")
	if err != nil {
		panic(err)
	}

	fmt.Println("Generating test configuration...")
	var buf bytes.Buffer
	err = tpl.ExecuteTemplate(&buf, "identity_provider.tmpl", testContext)
	if err != nil {
		panic(err)
	}
	result := buf.String()
	return result
}
