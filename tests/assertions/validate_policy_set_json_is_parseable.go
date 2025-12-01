package assertions

import (
	"context"
	"fmt"
	"github.com/cedar-policy/cedar-go"
	"github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func ValidatePolicySetJsonIsParseable(resourceName string) statecheck.StateCheck {
	return &validatePolicySetJsonIsParseable{
		resourceName: resourceName,
	}
}

type validatePolicySetJsonIsParseable struct {
	resourceName string
}

var _ statecheck.StateCheck = &validatePolicySetJsonIsParseable{}

func (e *validatePolicySetJsonIsParseable) CheckState(_ context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	if req.State == nil {
		resp.Error = fmt.Errorf("state is nil")

		return
	}

	if req.State.Values == nil {
		resp.Error = fmt.Errorf("state does not contain any state values")

		return
	}

	if req.State.Values.RootModule == nil {
		resp.Error = fmt.Errorf("state does not contain a root module")

		return
	}

	var rs *tfjson.StateResource
	for _, r := range req.State.Values.RootModule.Resources {
		if r.Address == e.resourceName {
			rs = r
			break
		}
	}

	if rs == nil {
		resp.Error = fmt.Errorf("not found: %s", e.resourceName)
		return
	}

	dataJson := rs.AttributeValues["data_json"].(string)

	if dataJson == "" {
		resp.Error = fmt.Errorf("data_cedar is not set")
		return
	}

	var unmarshalled cedar.Policy
	err := unmarshalled.UnmarshalJSON([]byte(dataJson))
	if err != nil {
		resp.Error = fmt.Errorf("failed to parse cedar schema JSON: %w", err)
		return
	}
}
