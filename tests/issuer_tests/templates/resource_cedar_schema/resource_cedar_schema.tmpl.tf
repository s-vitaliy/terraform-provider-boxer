provider "boxer" {
    external_auth = {
        security_token = "{{ .Token }}"
        identity_provider_id = "keycloak"
        internal_token_provider_endpoint = "http://localhost:5555/issuer"
    }

    issuer_host    = "http://localhost:5555/issuer"
    validator_host = "http://localhost:5555/validator"
}

resource "boxer_issuer_cedar_schema" "example" {
   id = "{{ .ObjectName }}"
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
