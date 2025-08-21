resource "boxer_action_discovery_document" "photo_app_api_v1" {
  id       = "photo-app-api-v1"
  schema   = boxer_validator_cedar_schema.integration_test.id
  hostname = "www.example.com"
  routes = [
    {
      method = "GET"
      path   = "api/v1/resources"
      action = "PhotoApp::Action::\"viewPhoto\""
    },
    {
      method = "GET"
      path   = "api/v2/resources"
      action = "PhotoApp::Action::\"viewPhoto\""
    },
  ]
}

resource "boxer_resource_discovery_document" "photo_app_api_v1_resources" {
  id       = "photo-app-api-v1"
  schema   = boxer_validator_cedar_schema.integration_test.id
  hostname = "www.example.com"
  routes = [
    {
      method   = "GET"
      path     = "api/v1/resources"
      resource = "PhotoApp::Photo::\"vacationPhoto.jpg\""
    },
    {
      method   = "GET"
      path     = "api/v2/resources"
      resource = "PhotoApp::Photo::\"vacationPhoto.jpg\""
    },
  ]
}
