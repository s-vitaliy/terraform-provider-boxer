package issuer

import (
	"context"
	"fmt"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &boxerPrincipal{}
	_ resource.ResourceWithConfigure = &boxerPrincipal{}
)

// NewBoxerPrincipalResource is a helper function to simplify the provider implementation.
func NewBoxerPrincipalResource() resource.Resource {
	return &boxerPrincipal{}
}

// boxerPrincipal is the resource implementation.
type boxerPrincipal struct {
	issuerClient *issuerClient.Client
}

func (resource *boxerPrincipal) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := getResourceIssuerClient(request, response)
	resource.issuerClient = client
}

// Metadata responds with the resource type name.
func (resource *boxerPrincipal) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_principal"
}

// Schema defines the schema for the resource.
func (resource *boxerPrincipal) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the principal.",
				Computed:    true,
			},
			"schema_id": schema.StringAttribute{
				Description: "The schema ID that this principal belongs to.",
				Required:    true,
			},
			"data_json": schema.StringAttribute{
				Description: "The principal data in JSON format.",
				Required:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (resource *boxerPrincipal) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var planModel boxerPrincipalModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	tflog.Info(ctx, "Creating Boxer Principal in schema", map[string]any{"schemaId": planModel.SchemaId.ValueString()})
	apiData, err := resource.issuerClient.PostPrincipal(ctx,
		jx.Raw(planModel.DataJson.ValueString()),
		issuerClient.PostPrincipalParams{Schema: planModel.SchemaId.ValueString()})

	if err != nil { // coverage-ignore
		common.GenerateError(&response.Diagnostics, "Creating", "Boxer Principal", err)
		return
	}

	err = saveNewBoxerPrincipalState(ctx,
		apiData.UID,
		planModel.DataJson.ValueString(),
		planModel.SchemaId.ValueString(),
		&response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil { // coverage-ignore
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (resource *boxerPrincipal) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var stateModel boxerPrincipalModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	// For now, we don't use the read result from the API since backend returns the normalized schema data
	// and if we use it, we will get a 'provider produced inconsistent result' error.
	// Instead, we just check if the schema exists and save the stateModel.
	// This will be updated in the future to use the read result.
	tflog.Info(ctx, "Getting principal by ID", map[string]any{"principalId": stateModel.ID.ValueString()})
	params := issuerClient.GetPrincipalParams{Schema: stateModel.SchemaId.ValueString(), ID: stateModel.ID.ValueString()}
	_, err = resource.issuerClient.GetPrincipal(ctx, params)
	if err != nil { // coverage-ignore
		common.GenerateError(&response.Diagnostics, "Reading", "Boxer Principal", err)
		return
	}
	err = saveNewBoxerPrincipalState(ctx,
		stateModel.ID.ValueString(),
		stateModel.DataJson.ValueString(),
		stateModel.SchemaId.ValueString(),
		&response.State, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't save the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (resource *boxerPrincipal) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var planModel boxerPrincipalModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the planModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	var stateModel boxerPrincipalModel
	err = common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	_, err = resource.issuerClient.PostPrincipal(ctx,
		jx.Raw(planModel.DataJson.ValueString()),
		issuerClient.PostPrincipalParams{Schema: planModel.SchemaId.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Updating", "Boxer Principal", err)
		return
	}

	err = saveNewBoxerPrincipalState(ctx,
		stateModel.ID.ValueString(),
		planModel.DataJson.ValueString(),
		stateModel.SchemaId.ValueString(),
		&response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil { // coverage-ignore
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (resource *boxerPrincipal) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var stateModel boxerPrincipalModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	/// We do not support deleting Principals directly in the Boxer API. The principals should be deleted only
	// together with the external identity associated with the principal.
	// Therefore, the method is empty. If user will try to delete the principal and then create the principal with the same ID,
	// the API will return an error that the principal with the same ID already exists. This is the expected behavior.

	// But we still need a way to remove the resource from the Terraform state.
}

type boxerPrincipalModel struct {
	ID       types.String `tfsdk:"id"`
	DataJson types.String `tfsdk:"data_json"`
	SchemaId types.String `tfsdk:"schema_id"`
}

func saveNewBoxerPrincipalState(ctx context.Context, id string, newData string, schemaId string, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	newState := boxerPrincipalModel{
		ID:       types.StringValue(id),
		DataJson: types.StringValue(newData),
		SchemaId: types.StringValue(schemaId),
	}
	diags := state.Set(ctx, &newState)
	diagnostics.Append(diags...)
	if diagnostics.HasError() { // coverage-ignore
		return fmt.Errorf("error getting plan")
	}
	return nil
}
