resource "boxer_issuer_cedar_schema" "integration_test" {
  id        = "integration-test-issuer"
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

resource "boxer_validator_cedar_schema" "integration_test" {
  id        = "integration-test-validator"
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

