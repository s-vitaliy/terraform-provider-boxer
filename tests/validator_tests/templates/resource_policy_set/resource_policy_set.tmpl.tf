provider "boxer" {
    external_auth = {
        security_token = "{{ .Token }}"
        identity_provider_id = "keycloak"
        internal_token_provider_endpoint = "http://localhost:5555/issuer"
    }

    issuer_host    = "http://localhost:5555/issuer"
    validator_host = "http://localhost:5555/validator"
}
resource "boxer_validator_cedar_schema" "example" {
  id        = "{{ .ObjectName }}"
  data_json = <<EOT
  {
    "PhotoApp": {
      "commonTypes": {
        "ContextType": {
          "type": "Record",
          "attributes": {
            "ip": {
                "type": "Extension",
                "name": "ipaddr",
                "required": false
            },
            "authenticated": {
                "type": "Boolean",
                "required": false
            }
          }
        }
      },
      "entityTypes": {
        "Photo": {
          "shape": {
            "type": "Record",
            "attributes": {
                "private": {
                  "type": "Boolean",
                  "required": true
              }
            }
          }
        }
      },
      "actions": {
        "viewPhoto": {
          "appliesTo": {
            "principalTypes": [
                "User"
            ],
            "resourceTypes": [
                "Photo"
            ],
            "context": {
                "type": "ContextType"
            }
          }
        }
      }
    }
  }
EOT
}

resource "boxer_policy_set" "example" {
  id         = "{{ .ObjectName }}"
  schema     = boxer_validator_cedar_schema.example.id
  data_cedar = <<EOT
  permit (
      principal == PhotoApp::User::"alice",
      action == PhotoApp::Action::"viewPhoto",
      resource == PhotoApp::Photo::"vacationPhoto.jpg"
  );
EOT
}
