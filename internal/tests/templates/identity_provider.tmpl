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
        "{{ .Services.ExternalIdp.Endpoint }}",
    ]

    audiences = [
        "account"
    ]
}
