package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"strings"
)

func GenerateError(diagnostics *diag.Diagnostics, operation string, object string, err error) {
	diagnostics.AddError(
		fmt.Sprintf("Error %s %s", operation, object),
		fmt.Sprintf("An error occurred while %s %s the identity provider: %s", strings.ToLower(operation), object, err.Error()),
	)
}

func ReadFromState(ctx context.Context, target any, baseState tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := baseState.Get(ctx, target)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting state")
	}
	return nil
}

func ReadFromPlan(ctx context.Context, target any, basePlan tfsdk.Plan, diagnostics *diag.Diagnostics) error {
	diags := basePlan.Get(ctx, target)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}

func ReadFromConfig(ctx context.Context, target interface{}, baseState tfsdk.Config, diagnostics *diag.Diagnostics) error {
	diags := baseState.Get(ctx, target)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting config")
	}
	return nil
}
