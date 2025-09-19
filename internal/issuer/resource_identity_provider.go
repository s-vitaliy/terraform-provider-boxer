package issuer

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &identityProviderResource{}
	_ resource.ResourceWithConfigure = &identityProviderResource{}
)

// NewIdentityProviderResource is a helper function to simplify the provider implementation.
func NewIdentityProviderResource() resource.Resource {
	return &identityProviderResource{}
}

func (resource *identityProviderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := getResourceIssuerClient(request, response)
	if client == nil {
		return
	}
	resource.issuerClient = client
}

// Metadata responds with the resource type name.
func (resource *identityProviderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_identity_provider"
}

// Schema defines the schema for the resource.
func (resource *identityProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the identity provider.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description:        "The name of the identity provider.",
				DeprecationMessage: "Use the id field instead",
				Optional:           true,
			},
			"discovery_url": schema.StringAttribute{
				Description: "The OIDC discovery URL of the identity provider.",
				Required:    true,
			},
			"user_id_claim": schema.StringAttribute{
				Description: "The claim used to identify the user in the identity provider's token.",
				Required:    true,
			},
			"audiences": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"issuers": schema.ListAttribute{
				Description: "List of issuers for the identity provider.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (resource *identityProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var planModel identityProviderResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	registration, diags := toBoxerIssuerModel(ctx, &planModel)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	err = resource.issuerClient.PostProvider(ctx, registration, issuerClient.PostProviderParams{ID: planModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Creating", "Identity Provider", err)
		return
	}
	planModel.ID = types.StringValue(planModel.Name.ValueString())
	response.State.Set(ctx, planModel)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (resource *identityProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var stateModel identityProviderResourceModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	apiData, err := resource.issuerClient.GetProvider(ctx, issuerClient.GetProviderParams{ID: stateModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Identity Provider", err)
		return
	}

	newStateModel := &identityProviderResourceModel{
		ID: stateModel.ID,
	}
	err = newStateModel.handleReadApiResponse(ctx, apiData, &response.State, &response.Diagnostics)
	if err != nil {
		// If we can't handle the API response, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (resource *identityProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var planModel identityProviderResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	var stateModel identityProviderResourceModel
	err = common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	registration, diags := toBoxerIssuerModel(ctx, &planModel)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	err = resource.issuerClient.PostProvider(ctx, registration, issuerClient.PostProviderParams{ID: stateModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Updating", "Identity Provider", err)
		return
	}
	apiData, err := resource.issuerClient.GetProvider(ctx, issuerClient.GetProviderParams{ID: stateModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Identity Provider", err)
		return
	}

	newStateModel := &identityProviderResourceModel{
		ID: stateModel.ID,
	}
	err = newStateModel.handleReadApiResponse(ctx, apiData, &response.State, &response.Diagnostics)
	if err != nil {
		// If we can't handle the API response, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (resource *identityProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var stateModel identityProviderResourceModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.issuerClient.DeleteProvider(ctx, issuerClient.DeleteProviderParams{ID: stateModel.ID.ValueString()})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Deleting", "Identity Provider", err)
		return
	}
}

func encode(values []string) []attr.Value {
	encoded := make([]attr.Value, len(values))
	for i, value := range values {
		encoded[i] = types.StringValue(value)
	}
	return encoded

}

func toBoxerIssuerModel(ctx context.Context, plan *identityProviderResourceModel) (*issuerClient.OidcIdentityProviderRegistration, diag.Diagnostics) {
	registration := issuerClient.OidcIdentityProviderRegistration{
		DiscoveryUrl: plan.DiscoveryUrl.ValueString(),
		UserIdClaim:  plan.UserIdClaim.ValueString(),
		Audiences:    make([]string, 0, len(plan.Audiences.Elements())),
		Issuers:      make([]string, len(plan.Issuers.Elements())),
	}
	diags := plan.Audiences.ElementsAs(ctx, &registration.Audiences, false)
	if diags.HasError() {
		return nil, diags
	}
	diags = plan.Issuers.ElementsAs(ctx, &registration.Issuers, false)
	if diags.HasError() {
		return nil, diags
	}
	return &registration, nil
}

func (model *identityProviderResourceModel) From(apiData *issuerClient.OidcIdentityProviderRegistration) *identityProviderResourceModel {
	model.Name = model.ID
	model.DiscoveryUrl = types.StringValue(apiData.GetDiscoveryUrl())
	model.UserIdClaim = types.StringValue(apiData.GetUserIdClaim())
	model.Audiences = types.ListValueMust(types.StringType, encode(apiData.GetAudiences()))
	model.Issuers = types.ListValueMust(types.StringType, encode(apiData.GetIssuers()))
	return model
}

func (model *identityProviderResourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}

func (model *identityProviderResourceModel) handleReadApiResponse(ctx context.Context, apiData issuerClient.GetProviderRes, state *tfsdk.State, diags *diag.Diagnostics) error {
	switch apiResponse := apiData.(type) {
	case *issuerClient.OidcIdentityProviderRegistration:
		// This is the expected type, we can proceed.
		err := model.From(apiResponse).saveToState(ctx, state, diags)
		if err != nil {
			// If we can't save the state, we can't proceed with the update.
			// so we return early.
			// The error will be handled by the framework and returned to the user.
			common.GenerateError(diags, "Saving", "Identity Provider State", err)
			return err
		}
	case *issuerClient.GetProviderNotFound:
		// If the API returns a not found error, we remove the resource from the state.
		state.RemoveResource(ctx)
		return nil
	default:
		// If the API returns an unexpected type, we generate an error.
		common.GenerateError(diags,
			"Reading",
			"Identity Provider",
			fmt.Errorf("unexpected type %T", apiData))
		return nil
	}
	return nil
}

type identityProviderResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DiscoveryUrl types.String `tfsdk:"discovery_url"`
	UserIdClaim  types.String `tfsdk:"user_id_claim"`
	Audiences    types.List   `tfsdk:"audiences"`
	Issuers      types.List   `tfsdk:"issuers"`
}

// identityProviderResource is the resource implementation.
type identityProviderResource struct {
	issuerClient *issuerClient.Client
}
