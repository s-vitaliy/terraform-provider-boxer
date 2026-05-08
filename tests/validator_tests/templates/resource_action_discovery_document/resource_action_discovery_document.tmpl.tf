provider "boxer" {
    external_auth = {
        security_token = "{{ .Token }}"
        identity_provider_id = "keycloak"
        internal_token_provider_endpoint = "http://localhost:5555/issuer"
    }

    issuer_host    = "http://localhost:5555/issuer"
    validator_host = "http://localhost:5555/validator"
}

resource "boxer_validator_cedar_schema" "integration_test" {
    id        = "{{ .ObjectName }}-validator"
    validate_data_json = true
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
              "User",
              "UserGroup"
            ],
            "resourceTypes": [
              "Photo"
            ],
            "context": {
              "type": "EntityOrCommon",
              "name": "ContextType"
            }
          }
        }
      }
    }
  }
EOT
}

resource "boxer_action_discovery_document" "example" {
    id       = "{{ .ObjectName }}"
    schema   = boxer_validator_cedar_schema.integration_test.id
    hostname = "www.example.com"
    routes = [
        {
            method = "GET"
            path   = "api/v1/resource"
            action = "PhotoApp::Action::\"viewPhoto\""
        },
        {
            method = "GET"
            path   = "api/v2/resource"
            action = "PhotoApp::Action::\"viewPhoto\""
        },
    ]
}
