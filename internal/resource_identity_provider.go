package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	issuer "terraform-provider-boxer/pkg/generated/api"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &identityProviderResource{}
	_ resource.ResourceWithConfigure = &identityProviderResource{}
)

// IdentityProviderResource is a helper function to simplify the provider implementation.
func IdentityProviderResource() resource.Resource {
	return &identityProviderResource{}
}

func (resource *identityProviderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	data, ok := request.ProviderData.(*BoxerProviderData)
	if !ok {
		response.Diagnostics.AddError(
			"Invalid Provider Data",
			"The provider data must be of type *BoxerProviderData, but was %s. This is most likely the bug in the provider implementation.",
		)
		return
	}
	if data.issuerClient == nil {
		response.Diagnostics.AddError(
			"Invalid Issuer Client",
			"The issuer client must not be nil. This is most likely the bug in the provider implementation.",
		)
		return
	}
	resource.issuerClient = data.issuerClient
}

// Metadata returns the resource type name.
func (resource *identityProviderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_identity_provider"
}

// Schema defines the schema for the resource.
func (resource *identityProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the identity provider.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the identity provider.",
				Required:    true,
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
	plan, err := readPlan(ctx, request.Plan, &response.Diagnostics)
	if err != nil {
		return
	}
	registration, diags := toBoxerIssuerModel(ctx, plan)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	err = resource.issuerClient.PostProvider(ctx, registration, issuer.PostProviderParams{ID: plan.Name.ValueString()})
	if err != nil {
		response.Diagnostics.AddError(
			"Error Creating Identity Provider",
			"An error occurred while creating the identity provider: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(plan.Name.ValueString())
	response.State.Set(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (resource *identityProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	state, err := readState(ctx, request.State, &response.Diagnostics)
	if err != nil {
		return
	}
	apiData, err := resource.issuerClient.GetProvider(ctx, issuer.GetProviderParams{ID: state.Name.ValueString()})
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading identity provider",
			"An error occurred while reading the identity provider: "+err.Error(),
		)
		return
	}
	newState := fromBoxerIssuerModel(state.ID.ValueString(), apiData)
	diags := response.State.Set(ctx, &newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (resource *identityProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan identityProviderResourceModel
	diags := request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var state identityProviderResourceModel
	diags = request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	registration, diags := toBoxerIssuerModel(ctx, &plan)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	err := resource.issuerClient.PostProvider(ctx, registration, issuer.PostProviderParams{ID: state.Name.ValueString()})
	if err != nil {
		response.Diagnostics.AddError(
			"Error Updating Identity Provider",
			"An error occurred while creating the identity provider: "+err.Error(),
		)
		return
	}
	apiData, err := resource.issuerClient.GetProvider(ctx, issuer.GetProviderParams{ID: state.Name.ValueString()})
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading identity provider",
			"An error occurred while reading the identity provider: "+err.Error(),
		)
		return
	}
	plan = fromBoxerIssuerModel(state.Name.ValueString(), apiData)
	diags = response.State.Set(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (resource *identityProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var plan identityProviderResourceModel
	diags := request.State.Get(ctx, &plan)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	err := resource.issuerClient.DeleteProvider(ctx, issuer.DeleteProviderParams{ID: plan.Name.ValueString()})
	if err != nil {
		response.Diagnostics.AddError(
			"Error Deleting Identity Provider",
			"An error occurred while deleting the identity provider: "+err.Error(),
		)
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

func toBoxerIssuerModel(ctx context.Context, plan *identityProviderResourceModel) (*issuer.OidcIdentityProviderRegistration, diag.Diagnostics) {
	registration := issuer.OidcIdentityProviderRegistration{
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

func fromBoxerIssuerModel(id string, apiData *issuer.OidcIdentityProviderRegistration) identityProviderResourceModel {
	return identityProviderResourceModel{
		ID:           types.StringValue(id),
		Name:         types.StringValue(id),
		DiscoveryUrl: types.StringValue(apiData.GetDiscoveryUrl()),
		UserIdClaim:  types.StringValue(apiData.GetUserIdClaim()),
		Audiences:    types.ListValueMust(types.StringType, encode(apiData.GetAudiences())),
		Issuers:      types.ListValueMust(types.StringType, encode(apiData.GetIssuers())),
	}
}

func readPlan(ctx context.Context, basePlan tfsdk.Plan, diagnostics *diag.Diagnostics) (*identityProviderResourceModel, error) {
	var plan identityProviderResourceModel
	diags := basePlan.Get(ctx, &plan)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("error getting plan")
	}
	return &plan, nil
}

func readState(ctx context.Context, baseState tfsdk.State, diagnostics *diag.Diagnostics) (*identityProviderResourceModel, error) {
	var state identityProviderResourceModel
	diags := baseState.Get(ctx, &state)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("error getting state")
	}
	return &state, nil
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
	issuerClient *issuer.Client
}
