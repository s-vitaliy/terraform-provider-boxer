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

data "boxer_identity_provider" "example" {
  id = "provider"
}

output "identity_provider" {
  value = data.boxer_identity_provider.example
}