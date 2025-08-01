package issuer

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-boxer/internal/common"
	issuer "terraform-provider-boxer/pkg/generated/api/issuerClient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &boxerExternalIdentity{}
	_ resource.ResourceWithConfigure = &boxerExternalIdentity{}
)

// NewBoxerExternalIdentityResource is a helper function to simplify the provider implementation.
func NewBoxerExternalIdentityResource() resource.Resource {
	return &boxerExternalIdentity{}
}

// boxerExternalIdentity is the resource implementation.
type boxerExternalIdentity struct {
	issuerClient *issuer.Client
}

func (resource *boxerExternalIdentity) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := getResourceIssuerClient(request, response)
	resource.issuerClient = client
}

// Metadata responds with the resource type name.
func (resource *boxerExternalIdentity) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_external_identity"
}

// Schema defines the schema for the resource.
func (resource *boxerExternalIdentity) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The unique identifier of the principal.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"identity_provider": schema.StringAttribute{
				Description:   "The identity provider that the external identity belongs to.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"principal": schema.SingleNestedAttribute{
				Description: "The principal ID associated to the external identity.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"principal_id": schema.StringAttribute{
						Description: "The unique identifier of the principal associated with the external identity.",
						Required:    true,
					},
					"schema_id": schema.StringAttribute{
						Description: "The schema ID of the principal associated with the external identity.",
						Required:    true,
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (resource *boxerExternalIdentity) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var planModel boxerExternalIdentityModel
	err := readFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	var associationPlanModel boxerPrincipalAssociationModel
	diags := request.Config.GetAttribute(ctx, path.Root("principal"), &associationPlanModel)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		// If we can't read the principal association model, we can't proceed.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	params := issuer.PostIdentityParams{
		IdentityProvider: planModel.IdentityProvider.ValueString(),
		ID:               planModel.ID.ValueString(),
	}
	err = resource.issuerClient.PostIdentity(ctx, params)
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Creating", "External Identity", err)
		return
	}

	associationRequest := issuer.IdentityAssociation{
		IdentityProvider: planModel.IdentityProvider.ValueString(),
		Identity:         planModel.ID.ValueString(),
		PrincipalID:      associationPlanModel.PrincipalId.ValueString(),
		PrincipalSchema:  associationPlanModel.SchemaId.ValueString(),
	}
	err = resource.issuerClient.PostAssociation(ctx, &associationRequest)
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Creating", "External Identity to Principal Association", err)
		return
	}

	err = saveNewExternalIdentityState(ctx, &planModel, &response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (resource *boxerExternalIdentity) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var stateModel boxerExternalIdentityModel
	err := readFromState(ctx, &stateModel, request.State, &response.Diagnostics)
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
	tflog.Info(ctx, "Getting external identity by ID", map[string]any{"principalId": stateModel.ID.ValueString()})
	externalId, err := resource.issuerClient.GetIdentity(ctx, issuer.GetIdentityParams{
		IdentityProvider: stateModel.IdentityProvider.ValueString(),
		ID:               stateModel.ID.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "External Identity", err)
		return
	}

	association, err := resource.issuerClient.GetAssociation(ctx, issuer.GetAssociationParams{
		IdentityProvider: stateModel.IdentityProvider.ValueString(),
		ID:               stateModel.ID.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "External Identity association", err)
		return
	}

	model := boxerExternalIdentityModel{
		ID:               types.StringValue(externalId.UserId),
		IdentityProvider: types.StringValue(externalId.IdentityProvider),
		Principal: boxerPrincipalAssociationModel{
			PrincipalId: types.StringValue(association.PrincipalID),
			SchemaId:    types.StringValue(association.PrincipalSchema),
		},
	}

	err = saveNewExternalIdentityState(ctx, &model, &response.State, &response.Diagnostics)
	if err != nil {
		// If we can't save the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (resource *boxerExternalIdentity) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var planModel boxerExternalIdentityModel
	err := readFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the planModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	err = resource.issuerClient.PostIdentity(ctx, issuer.PostIdentityParams{
		IdentityProvider: planModel.IdentityProvider.ValueString(),
		ID:               planModel.ID.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Updating", "External Identity", err)
		return
	}

	var associationPlanModel boxerPrincipalAssociationModel
	diags := request.Config.GetAttribute(ctx, path.Root("principal"), &associationPlanModel)
	response.Diagnostics.Append(diags...)
	if diags.HasError() {
		// If we can't read the principal association model, we can't proceed.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.issuerClient.PostAssociation(ctx, &issuer.IdentityAssociation{
		IdentityProvider: planModel.IdentityProvider.ValueString(),
		Identity:         planModel.ID.ValueString(),
		PrincipalID:      associationPlanModel.PrincipalId.ValueString(),
		PrincipalSchema:  associationPlanModel.SchemaId.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Updating", "External Identity association", err)
		return
	}

	err = saveNewExternalIdentityState(ctx, &planModel, &response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (resource *boxerExternalIdentity) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var stateModel boxerExternalIdentityModel
	err := readFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	tflog.Info(ctx, "Deleting external identity", map[string]any{"identityProvider": stateModel.IdentityProvider.ValueString(), "id": stateModel.ID.ValueString()})
	err = resource.issuerClient.DeleteIdentity(ctx, issuer.DeleteIdentityParams{
		IdentityProvider: stateModel.IdentityProvider.ValueString(),
		ID:               stateModel.ID.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Deleting", "External Identity", err)
		return
	}
}

type boxerExternalIdentityModel struct {
	ID               types.String                   `tfsdk:"id"`
	IdentityProvider types.String                   `tfsdk:"identity_provider"`
	Principal        boxerPrincipalAssociationModel `tfsdk:"principal"`
}

type boxerPrincipalAssociationModel struct {
	PrincipalId types.String `tfsdk:"principal_id"`
	SchemaId    types.String `tfsdk:"schema_id"`
}

func saveNewExternalIdentityState(ctx context.Context, newState *boxerExternalIdentityModel, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, newState)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}
