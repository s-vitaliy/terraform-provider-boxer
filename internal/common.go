package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"strings"
)

func generateError(diagnostics *diag.Diagnostics, operation string, object string, err error) {
	diagnostics.AddError(
		fmt.Sprintf("Error %s %s", operation, object),
		fmt.Sprintf("An error occurred while %s %s the identity provider: %s", strings.ToLower(operation), object, err.Error()),
	)
}
