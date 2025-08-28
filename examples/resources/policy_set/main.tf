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

resource "boxer_policy_set" "example" {
  id         = "example"
  data_cedar = <<EOT
  permit (
      principal == PhotoApp::User::"alice",
      action == PhotoApp::Action::"viewPhoto",
      resource == PhotoApp::Photo::"vacationPhoto.jpg"
  );

  permit (
      principal == PhotoApp::User::"stacey",
      action == PhotoApp::Action::"viewPhoto",
      resource
  )
  when { resource in PhotoApp::Account::"stacey" };
EOT
}

output "test" {
  value = boxer_policy_set.example
}