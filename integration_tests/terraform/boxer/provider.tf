terraform {
  required_providers {
    boxer = {
      source = "registry.terraform.io/sneaksAndData/boxer"
    }
  }
}

provider "boxer" {
  external_auth = {
    security_token = var.external_token
    identity_provider_id = "root"
    internal_token_provider_endpoint = "http://localhost:8888/"
  }

  issuer_host = "http://localhost:8888/"
  validator_host = "http://localhost:8081/"
}
