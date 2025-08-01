package pkg

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target generated/api/issuerClient -package issuerClient --clean ./openapi.issuer.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target generated/api/validatorClient -package validatorClient --clean ./openapi.validator.json
