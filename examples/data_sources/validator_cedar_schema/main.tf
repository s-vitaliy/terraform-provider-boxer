terraform {
  required_providers {
    boxer = {
      source = "registry.terraform.io/sneaksAndData/boxer"
    }
  }
}

provider "boxer" {
  issuer_host    = "http://localhost:8888/"
  validator_host = "http://localhost:8081/"
}

resource "boxer_validator_cedar_schema" "example" {
  id                 = "example"
  validate_data_json = true
  data_json          = <<EOT
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
                      "required": true
                  }
              }
          }
      },
      "entityTypes": {
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
            },
            "createPhoto": {
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
            },
            "listPhotos": {
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

data "boxer_validator_cedar_schema" "example" {
  id = boxer_validator_cedar_schema.example.id
}

output "test" {
  value = data.boxer_validator_cedar_schema.example
}
