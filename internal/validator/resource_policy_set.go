package validator

import (
	"context"
	"fmt"
	"github.com/cedar-policy/cedar-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &policySetResource{}
	_ resource.ResourceWithConfigure        = &policySetResource{}
	_ resource.ResourceWithConfigValidators = &policySetResource{}
)

// NewPolicySetResource is a helper function to simplify the provider implementation.
func NewPolicySetResource() resource.Resource {
	return &policySetResource{}
}

// policySetResource is the resource implementation.
type policySetResource struct {
	validatorClient *validatorClient.Client
}

func (r *policySetResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("data_cedar"),
			path.MatchRoot("data_json"),
		),
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("data_cedar"),
			path.MatchRoot("data_json"),
		),
	}
}

func (r *policySetResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := getResourceValidatorClient(request, response)
	r.validatorClient = client
}

// Metadata responds with the resource type name.
func (r *policySetResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_policy_set"
}

// Schema defines the schema for the resource.
func (r *policySetResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the policy set.",
				Required:    true,
			},
			"data_cedar": schema.StringAttribute{
				Description: "The Cedar schema data in Cedar format.",
				Optional:    true,
				Computed:    true,
			},
			"data_json": schema.StringAttribute{
				Description: "The Cedar schema data in Cedar format.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *policySetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var planModel policySetResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = planModel.normalize()
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Normalizing", "Policy Set", err)
		return
	}

	err = r.validatorClient.PostPolicySet(ctx, planModel.Into(), validatorClient.PostPolicySetParams{ID: planModel.ID.ValueString()})

	if err != nil {
		common.GenerateError(&response.Diagnostics, "Creating", "Policy Set", err)
		return
	}

	err = planModel.saveToState(ctx, &response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *policySetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var stateModel policySetResourceModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = stateModel.normalize()
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Normalizing", "Policy Set", err)
		return
	}
	// For now, we don't use the read result from the API since backend returns the normalized schema data
	// and if we use it, we will get a 'provider produced inconsistent result' error.
	// Instead, we just check if the schema exists and save the stateModel.
	// This will be updated in the future to use the read result.
	registration, err := r.validatorClient.GetPolicySet(ctx, validatorClient.GetPolicySetParams{ID: stateModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Policy Set", err)
		return
	}

	apiModel := &policySetResourceModel{
		ID: stateModel.ID,
	}

	apiModel, err = apiModel.From(registration)
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Converting", "Policy Set", err)
	}
	err = apiModel.saveToState(ctx, &response.State, &response.Diagnostics)
	// If we can't save the stateModel, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *policySetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var planModel policySetResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the planModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = planModel.normalize()
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Normalizing", "Policy Set", err)
	}
	err = r.validatorClient.PostPolicySet(ctx, planModel.Into(), validatorClient.PostPolicySetParams{ID: planModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Updating", "Policy Set", err)
		return
	}

	err = planModel.saveToState(ctx, &response.State, &response.Diagnostics)
	// If we can't save the stateModel, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *policySetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var stateModel policySetResourceModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	err = r.validatorClient.DeletePolicySet(ctx, validatorClient.DeletePolicySetParams{ID: stateModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Deleting", "Policy Set", err)
		return
	}
}

type policySetResourceModel struct {
	ID        types.String `tfsdk:"id"`
	DataCedar types.String `tfsdk:"data_cedar"`
	DataJson  types.String `tfsdk:"data_json"`
}

func (model *policySetResourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}

func (model *policySetResourceModel) Into() *validatorClient.PolicySetRegistration {
	registration := validatorClient.PolicySetRegistration{
		Policy: model.DataCedar.ValueString(),
	}

	return &registration
}

func (model *policySetResourceModel) From(source *validatorClient.PolicySetRegistration) (*policySetResourceModel, error) {
	model.DataCedar = types.StringValue(source.Policy)
	err := model.normalize()
	if err != nil {
		return nil, fmt.Errorf("error normalizing policy set model: %w", err)
	}
	return model, nil
}

func (model *policySetResourceModel) normalize() error {
	var policy cedar.Policy

	if model.DataJson.IsUnknown() {
		if model.DataCedar.IsUnknown() || model.DataCedar.IsNull() {
			return fmt.Errorf("either data_cedar or data_json must be set")
		}

		err := policy.UnmarshalCedar([]byte(model.DataCedar.ValueString()))
		if err != nil {
			return fmt.Errorf("error unmarshalling cedar policy: %w", err)
		}
		jsonValue, err := policy.MarshalJSON()
		if err != nil {
			return fmt.Errorf("error marshalling cedar policy to json: %w", err)
		}
		model.DataJson = types.StringValue(string(jsonValue))
		return nil
	}

	if model.DataCedar.IsUnknown() {
		if model.DataJson.IsUnknown() || model.DataJson.IsNull() {
			return fmt.Errorf("either data_cedar or data_json must be set")
		}

		err := policy.UnmarshalJSON([]byte(model.DataJson.ValueString()))
		if err != nil {
			return fmt.Errorf("error unmarshalling cedar policy: %w", err)
		}
		cedarValue := policy.MarshalCedar()
		model.DataCedar = types.StringValue(string(cedarValue))
		return nil
	}

	if model.DataCedar.IsNull() && model.DataJson.IsNull() || model.DataCedar.IsUnknown() && model.DataJson.IsUnknown() {
		return fmt.Errorf("either data_cedar or data_json must be set")
	}

	return nil
}
