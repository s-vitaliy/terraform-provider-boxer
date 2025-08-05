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

resource "boxer_action_discovery_document" "example" {
  id       = "example"
  hostname = "www.example.com"
  routes = [
    {
      method = "GET"
      path   = "/api/v1/resources"
      action = "resourceManager::ListResources"
    },
    {
      method = "POST"
      path   = "/api/v1/resources"
      action = "resourceManager::CreateResource"
    },
    {
      method = "GET"
      path   = "/api/v1/resources/{id}"
      action = "resourceManager::GetResource"
    },
    {
      method = "PUT"
      path   = "/api/v1/resources/{id}"
      action = "resourceManager::UpdateResource"
    },
    {
      method = "DELETE"
      path   = "/api/v1/resources/{id}"
      action = "resourceManager::DeleteResource"
    },
    {
      method = "GET"
      path   = "/api/v1/users"
      action = "identityService::ListUsers"
    },
  ]
}

data "boxer_action_discovery_document" "example" {
  id = boxer_action_discovery_document.example.id
}

output "test" {
  value = data.boxer_action_discovery_document.example
}