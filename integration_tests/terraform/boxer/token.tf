data boxer_token "example" {
  identity_provider = boxer_external_identity.alice.identity_provider
  auth = {
    # For testing purposes, provide the bearer token value manually
    bearer = var.external_token
  }
}
