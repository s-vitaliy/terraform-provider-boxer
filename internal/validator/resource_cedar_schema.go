package validator

import (
	"context"
	"fmt"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &cedarSchemaResource{}
	_ resource.ResourceWithConfigure = &cedarSchemaResource{}
)

// NewCedarSchemaResource is a helper function to simplify the provider implementation.
func NewCedarSchemaResource() resource.Resource {
	return &cedarSchemaResource{}
}

// cedarSchemaResource is the resource implementation.
type cedarSchemaResource struct {
	validatorClient *validatorClient.Client
}

func (resource *cedarSchemaResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := getResourceValidatorClient(request, response)
	resource.validatorClient = client
}

// Metadata responds with the resource type name.
func (resource *cedarSchemaResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_validator_cedar_schema"
}

// Schema defines the schema for the resource.
func (resource *cedarSchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the identity provider.",
				Required:    true,
			},
			"data_json": schema.StringAttribute{
				Description: "The schema data in JSON format.",
				Required:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (resource *cedarSchemaResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var planModel cedarSchemaResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.validatorClient.PostSchema(ctx, jx.Raw(planModel.DataJson.ValueString()), validatorClient.PostSchemaParams{ID: planModel.ID.ValueString()})

	if err != nil {
		common.GenerateError(&response.Diagnostics, "Creating", "Cedar Schema", err)
		return
	}

	err = saveNewState(ctx, planModel.ID.ValueString(), planModel.DataJson.ValueString(), &response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (resource *cedarSchemaResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var stateModel cedarSchemaResourceModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	// For now, we don't use the read result from the API since backend returns the normalized schema data
	// and if we use it, we will get a 'provider produced inconsistent result' error.
	// Instead, we just check if the schema exists and save the stateModel.
	// This will be updated in the future to use the read result.
	_, err = resource.validatorClient.GetSchema(ctx, validatorClient.GetSchemaParams{ID: stateModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Cedar Schema", err)
		return
	}
	err = saveNewState(ctx, stateModel.ID.ValueString(), stateModel.DataJson.ValueString(), &response.State, &response.Diagnostics)
	// If we can't save the stateModel, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (resource *cedarSchemaResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var planModel cedarSchemaResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the planModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	var stateModel cedarSchemaResourceModel
	err = common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.validatorClient.PostSchema(ctx, jx.Raw(planModel.DataJson.ValueString()), validatorClient.PostSchemaParams{ID: planModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Updating", "Cedar Schema", err)
		return
	}

	err = saveNewState(ctx, stateModel.ID.ValueString(), planModel.DataJson.ValueString(), &response.State, &response.Diagnostics)
	// If we can't save the stateModel, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (resource *cedarSchemaResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var stateModel cedarSchemaResourceModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	err = resource.validatorClient.DeleteSchema(ctx, validatorClient.DeleteSchemaParams{ID: stateModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Deleting", "Cedar Schema", err)
		return
	}
}

type cedarSchemaResourceModel struct {
	ID       types.String `tfsdk:"id"`
	DataJson types.String `tfsdk:"data_json"`
}

func saveNewState(ctx context.Context, id string, newData string, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	newState := cedarSchemaResourceModel{
		ID:       types.StringValue(id),
		DataJson: types.StringValue(newData),
	}
	diags := state.Set(ctx, &newState)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}
