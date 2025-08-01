terraform {
  required_providers {
    boxer = {
      source = "registry.terraform.io/sneaksAndData/boxer"
    }
  }
}

provider "boxer" {
  issuer_host = "http://localhost:8888/"
}

resource "boxer_cedar_schema" "example"  {
  id = "example"
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
          "memberOfTypes": [
            "UserGroup"
          ],
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
        },
        "UserGroup": {}
      },
      "actions": {}
    }
  }
EOT
}

resource "boxer_principal" "example" {
  schema_id = boxer_cedar_schema.example.id
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
    "parents": [
        {
            "type": "PhotoApp::UserGroup",
            "id": "alice_friends"
        },
        {
            "type": "PhotoApp::UserGroup",
            "id": "AVTeam"
        }
    ]
}
EOT
}

resource "boxer_identity_provider" "example"  {
  name = "provider"
  user_id_claim = "preferred_username"
  discovery_url = "http://localhost:8080/realms/master/"
  issuers = [
    "http://localhost:8080/realms/master",
  ]
  audiences = [
    "account"
  ]
}

resource "boxer_external_identity" "example" {
  identity_provider = "provider"
  id                   = "test_user"
  principal = {
    schema_id = boxer_cedar_schema.example.id
    principal_id = boxer_principal.example.id
  }
}

data boxer_token "example" {
  identity_provider = boxer_identity_provider.example.name
  auth = {
    # For testing purposes, provide the bearer token value manually
    bearer = ""
  }
}

output "test" {
  value = data.boxer_token.example
}
