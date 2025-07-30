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

data "boxer_cedar_schema" "test" {
  id = "test"
}

output "test" {
  value = data.boxer_cedar_schema.test
}