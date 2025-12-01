provider "boxer" {
    external_auth = {
        security_token = "{{ .Token }}"
        identity_provider_id = "keycloak"
        internal_token_provider_endpoint = "http://localhost:5555/issuer"
    }

    issuer_host    = "http://localhost:5555/issuer"
    validator_host = "http://localhost:5555/validator"
}

resource "boxer_identity_provider" "example" {
  id = "{{ .ObjectName }}"
  user_id_claim = "preferred_username"
  discovery_url = "{{ .Services.ExternalIdp.ClusterEndpoint }}"
  issuers = [
    "{{ .Services.ExternalIdp.Endpoint }}",
  ]

  audiences = [
    "account"
  ]
}

resource "boxer_issuer_cedar_schema" "example" {
   id = "{{ .ObjectName }}-issuer"
      data_json = <<EOT
      {
        "PhotoApp": {
          "commonTypes": {
            "PersonType": {
              "type": "Record",
              "attributes": {
                "age": {
                  "type": "Long"
                },
                "name": {
                  "type": "String"
                }
              }
            }
          },
          "entityTypes": {
            "User": {
              "shape": {
                "type": "Record",
                "attributes": {
                  "personInformation": {
                    "type": "PersonType"
                  },
                  "userId": {
                    "type": "String"
                  }
                }
              }
            }
          },
          "actions": {}
        }
      }
EOT
}

resource "boxer_validator_cedar_schema" "example" {
  id        = "{{ .ObjectName }}-validator"
  data_json = <<EOT
  {
    "PhotoApp": {
      "commonTypes": { },
      "entityTypes": { },
      "actions": { }
    }
  }
EOT
}



resource "boxer_principal" "example" {
  schema_id = boxer_issuer_cedar_schema.example.id
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


resource "boxer_external_identity" "example" {
  identity_provider = boxer_identity_provider.example.id
  id                = "{{ .ObjectName }}"
  principal = {
    schema_id    = boxer_principal.example.schema_id
    principal_id = boxer_principal.example.id
  }
  validator_schema_id = boxer_validator_cedar_schema.example.id
}

data "boxer_external_identity" "example" {
  identity_provider = boxer_identity_provider.example.id
  id                = boxer_external_identity.example.id
}
