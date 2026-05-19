# terraform-provider-boxer

A Terraform provider for managing [Boxer](https://github.com/SneaksAndData/boxer-issuer) resources.

## Authentication to the Keycloak server for the manual testing of the provider

```shell
$ curl \
  -d "client_id=test_client" \
  -d "client_secret=test_client_secret" \
  -d "username=test_root" \
  -d "password=test-root-password" \
  -d "grant_type=password" \
  "http://localhost:5555/auth/realms/master/protocol/openid-connect/token"
```