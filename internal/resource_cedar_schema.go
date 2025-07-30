package provider

import (
	"context"
	"fmt"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	issuer "terraform-provider-boxer/pkg/generated/api"

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
	issuerClient *issuer.Client
}

func (resource *cedarSchemaResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := getResourceIssuerClient(request, response)
	resource.issuerClient = client
}

// Metadata responds with the resource type name.
func (resource *cedarSchemaResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cedar_schema"
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
	plan, err := readCedarSchemaPlan(ctx, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.issuerClient.PostSchema(ctx, jx.Raw(plan.DataJson.ValueString()), issuer.PostSchemaParams{ID: plan.ID.ValueString()})

	if err != nil {
		generateError(&response.Diagnostics, "Creating", "Cedar Schema", err)
		return
	}

	err = saveNewState(ctx, plan.ID.ValueString(), plan.DataJson.ValueString(), &response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (resource *cedarSchemaResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	state, err := readCedarSchemaState(ctx, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	// For now, we don't use the read result from the API since backend returns the normalized schema data
	// and if we use it, we will get a 'provider produced inconsistent result' error.
	// Instead, we just check if the schema exists and save the state.
	// This will be updated in the future to use the read result.
	_, err = resource.issuerClient.GetSchema(ctx, issuer.GetSchemaParams{ID: state.ID.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Reading", "Cedar Schema", err)
		return
	}
	err = saveNewState(ctx, state.ID.ValueString(), state.DataJson.ValueString(), &response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (resource *cedarSchemaResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	plan, err := readCedarSchemaPlan(ctx, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the plan, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	state, err := readCedarSchemaState(ctx, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.issuerClient.PostSchema(ctx, jx.Raw(plan.DataJson.ValueString()), issuer.PostSchemaParams{ID: plan.ID.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Updating", "Cedar Schema", err)
		return
	}

	err = saveNewState(ctx, state.ID.ValueString(), plan.DataJson.ValueString(), &response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (resource *cedarSchemaResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	plan, err := readCedarSchemaState(ctx, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	err = resource.issuerClient.DeleteSchema(ctx, issuer.DeleteSchemaParams{ID: plan.ID.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Deleting", "Cedar Schema", err)
		return
	}
}

type cedarSchemaResourceModel struct {
	ID       types.String `tfsdk:"id"`
	DataJson types.String `tfsdk:"data_json"`
}

func readCedarSchemaState(ctx context.Context, baseState tfsdk.State, diagnostics *diag.Diagnostics) (*cedarSchemaResourceModel, error) {
	var state cedarSchemaResourceModel
	diags := baseState.Get(ctx, &state)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("error getting state")
	}
	return &state, nil
}

func readCedarSchemaPlan(ctx context.Context, basePlan tfsdk.Plan, diagnostics *diag.Diagnostics) (*cedarSchemaResourceModel, error) {
	var plan cedarSchemaResourceModel
	diags := basePlan.Get(ctx, &plan)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("error getting plan")
	}
	return &plan, nil
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
