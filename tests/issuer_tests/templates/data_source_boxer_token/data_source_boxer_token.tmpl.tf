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
    "{{ .Services.ExternalIdp.Endpoint }}/realsms/master",
  ]

  audiences = [
    "account"
  ]
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

resource "boxer_validator_cedar_schema" "example" {
  id        = "{{ .ObjectName }}-validator"
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



resource "boxer_principal" "example" {
  schema_id = boxer_issuer_cedar_schema.example.id
  data_json = <<EOT
{
    "uid": {
        "type": "PhotoApp::User",
        "id": "{{ .ObjectName }}-principal"
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
  id                = "test_user"
  principal = {
    schema_id    = boxer_principal.example.schema_id
    principal_id = boxer_principal.example.id
  }
  validator_schema_id = boxer_validator_cedar_schema.example.id
}


resource "boxer_action_discovery_document" "example" {
  id       = "{{ .ObjectName }}-action-discovery"
  schema   = boxer_validator_cedar_schema.example.id
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

resource "boxer_resource_discovery_document" "example" {
  id       = "{{ .ObjectName }}-resource-discovery"
  schema   = boxer_validator_cedar_schema.example.id
  hostname = "www.example.com"
  routes = [
    {
      method   = "GET"
      path     = "api/v1/resource"
      resource = "PhotoApp::Photo::\"vacationPhoto.jpg\""
    },
    {
      method   = "GET"
      path     = "api/v2/resource"
      resource = "PhotoApp::Photo::\"vacationPhoto.jpg\""
    },
  ]
}

resource "boxer_policy_set" "vacation_photo_access_policy" {
  id         = "vacation-photo-access-policy"
  schema     = boxer_validator_cedar_schema.example.id
  data_cedar = <<EOT
  permit (
      principal == PhotoApp::User::"alice",
      action == PhotoApp::Action::"viewPhoto",
      resource == PhotoApp::Photo::"vacationPhoto.jpg"
  );
EOT
}

data boxer_token "example" {
  identity_provider = "keycloak"
  auth = {
    bearer = "{{ .Token }}"
  }
}


