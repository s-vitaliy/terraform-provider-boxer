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

resource "boxer_resource_discovery_document" "example" {
  id       = "example"
  hostname = "www.example.com"
  routes = [
    {
      path   = "/api/v1/resources"
      resource = "resourceManager::ResourceList"
    },
    {
      path   = "/api/v1/resources/{id}"
      resource = "resourceManager::ResourceDetails"
    },
    {
      path   = "/api/v1/users"
      resource = "identityService::UserList"
    },
  ]
}

data "boxer_resource_discovery_document" "example" {
  id = boxer_resource_discovery_document.example.id
}

output "test" {
  value = data.boxer_resource_discovery_document.example
}