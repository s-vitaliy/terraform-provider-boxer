resource "boxer_identity_provider" "keycloak" {
  name          = "keycloak"
  user_id_claim = "preferred_username"
  discovery_url = "http://localhost:8080/realms/master/"
  issuers = [
    "http://localhost:8080/realms/master",
  ]
  audiences = [
    "account"
  ]
}

resource "boxer_principal" "alice" {
  schema_id = boxer_issuer_cedar_schema.integration_test.id
  data_json = <<EOT
{
    "uid": {
        "type": "PhotoApp::User",
        "id": "alice"
    },
    "attrs": {
        "userId": "897345789237492878",
        "personInformation": {
            "age": 85,
            "name": "alice"
        }
    },
    "parents": [ ]
}
EOT
}

resource "boxer_external_identity" "alice" {
  identity_provider = boxer_identity_provider.keycloak.name
  id                = "test_user"
  principal = {
    schema_id    = boxer_principal.alice.schema_id
    principal_id = boxer_principal.alice.id
  }
  validator_schema_id = boxer_validator_cedar_schema.integration_test.id
}

