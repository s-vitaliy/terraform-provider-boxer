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
	plan, err := readIdentityProviderPlan(ctx, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	registration, diags := toBoxerIssuerModel(ctx, plan)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	err = resource.issuerClient.PostProvider(ctx, registration, issuer.PostProviderParams{ID: plan.Name.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Creating", "Identity Provider", err)
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
	state, err := readIdentityProviderState(ctx, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	apiData, err := resource.issuerClient.GetProvider(ctx, issuer.GetProviderParams{ID: state.Name.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Reading", "Identity Provider", err)
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
	plan, err := readIdentityProviderPlan(ctx, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	state, err := readIdentityProviderState(ctx, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	registration, diags := toBoxerIssuerModel(ctx, plan)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	err = resource.issuerClient.PostProvider(ctx, registration, issuer.PostProviderParams{ID: state.Name.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Updating", "Identity Provider", err)
		return
	}
	apiData, err := resource.issuerClient.GetProvider(ctx, issuer.GetProviderParams{ID: state.Name.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Reading", "Identity Provider", err)
		return
	}
	newState := fromBoxerIssuerModel(state.Name.ValueString(), apiData)
	diags = response.State.Set(ctx, newState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (resource *identityProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	state, err := readIdentityProviderState(ctx, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.issuerClient.DeleteProvider(ctx, issuer.DeleteProviderParams{ID: state.Name.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Deleting", "Identity Provider", err)
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

func readIdentityProviderState(ctx context.Context, baseState tfsdk.State, diagnostics *diag.Diagnostics) (*identityProviderResourceModel, error) {
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

func readIdentityProviderPlan(ctx context.Context, basePlan tfsdk.Plan, diagnostics *diag.Diagnostics) (*identityProviderResourceModel, error) {
	var plan identityProviderResourceModel
	diags := basePlan.Get(ctx, &plan)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return nil, fmt.Errorf("error getting plan")
	}
	return &plan, nil
}
