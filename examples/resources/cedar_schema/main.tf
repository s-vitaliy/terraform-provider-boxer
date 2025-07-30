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

output "test" {
  value = boxer_cedar_schema.example
}