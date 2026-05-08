provider "boxer" {
    external_auth = {
        security_token = "{{ .Token }}"
        identity_provider_id = "keycloak"
        internal_token_provider_endpoint = "http://localhost:5555/issuer"
    }

    issuer_host    = "http://localhost:5555/issuer"
    validator_host = "http://localhost:5555/validator"
}
# The data_json is invalid in this case, as ContextType is not defined
resource "boxer_issuer_cedar_schema" "example" {
   id = "{{ .ObjectName }}"
   validate_data_json = true
   data_json = <<EOT
{
  "PhotoApp": {
    "commonTypes": {
      "PersonType": {
        "type": "Record",
        "attributes": {
          "age": { "type": "Long" },
          "name": { "type": "String" }
        }
      }
    },
    "entityTypes": {
      "User": {
        "shape": {
          "type": "Record",
          "attributes": {
            "personInformation": {
              "type": "EntityOrCommon",
              "name": "PersonType"
            },
            "userId": { "type": "String" }
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
            "type": "ContextType"
          }
        }
      }
    }
  }
}
EOT
}
