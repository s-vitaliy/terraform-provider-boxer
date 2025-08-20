package validator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &resourceDiscoveryDocumentResource{}
	_ resource.ResourceWithConfigure = &resourceDiscoveryDocumentResource{}
)

// NewResourceDiscoveryDocumentResource is a helper function to simplify the provider implementation.
func NewResourceDiscoveryDocumentResource() resource.Resource {
	return &resourceDiscoveryDocumentResource{}
}

// resourceDiscoveryDocumentResource is the resource implementation.
type resourceDiscoveryDocumentResource struct {
	validatorClient *validatorClient.Client
}

func (resource *resourceDiscoveryDocumentResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := getResourceValidatorClient(request, response)
	resource.validatorClient = client
}

// Metadata responds with the resource type name.
func (resource *resourceDiscoveryDocumentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_resource_discovery_document"
}

// Schema defines the schema for the resource.
func (resource *resourceDiscoveryDocumentResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the resource discovery document.",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "The schema that the action discovery document belongs to.",
				Required:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the resource discovery document.",
				Required:    true,
			},
			"routes": schema.ListNestedAttribute{
				Description: "The list of routes for the resource discovery document.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Description: "The path for the route.",
							Required:    true,
						},
						"resource": schema.StringAttribute{
							Description: "The resource associated with the route.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (resource *resourceDiscoveryDocumentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var planModel resourceDiscoveryDocumentResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.validatorClient.PostResourceSet(ctx, planModel.Into(), validatorClient.PostResourceSetParams{
		ID:     planModel.ID.ValueString(),
		Schema: planModel.Schema.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Creating", "Resource Set", err)
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
func (resource *resourceDiscoveryDocumentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var stateModel resourceDiscoveryDocumentResourceModel
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
	apiData, err := resource.validatorClient.GetResourceSet(ctx, validatorClient.GetResourceSetParams{
		ID:     stateModel.ID.ValueString(),
		Schema: stateModel.Schema.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Resource Set", err)
		return
	}

	apiModel := &resourceDiscoveryDocumentResourceModel{
		ID:     stateModel.ID,
		Schema: stateModel.Schema,
	}
	switch apiResponse := apiData.(type) {
	case *validatorClient.ResourceSetRegistration:
		err = apiModel.From(apiResponse).saveToState(ctx, &response.State, &response.Diagnostics)
		// If we can't save the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		if err != nil {
			return
		}
		return
	case *validatorClient.GetResourceSetNotFound:
		// If the resource set is not found, we remove the resource from the state.
		tflog.Debug(ctx, "Resource set not found, setting state to empty")
		response.State.RemoveResource(ctx)
		return
	default:
		common.GenerateError(&response.Diagnostics,
			"Reading",
			"Resource Set",
			common.ErrUnexpectedResponseType(apiResponse))
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (resource *resourceDiscoveryDocumentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var planModel resourceDiscoveryDocumentResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil {
		// If we can't read the planModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.validatorClient.PostResourceSet(ctx, planModel.Into(), validatorClient.PostResourceSetParams{
		ID:     planModel.ID.ValueString(),
		Schema: planModel.Schema.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Updating", "Resource Set", err)
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
func (resource *resourceDiscoveryDocumentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var stateModel resourceDiscoveryDocumentResourceModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)
	if err != nil {
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.validatorClient.DeleteResourceSet(ctx, validatorClient.DeleteResourceSetParams{
		ID:     stateModel.ID.ValueString(),
		Schema: stateModel.Schema.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Deleting", "Resource Set", err)
		return
	}
}

type resourceResourceRouteModel struct {
	Path     types.String `tfsdk:"path"`
	Resource types.String `tfsdk:"resource"`
}

type resourceDiscoveryDocumentResourceModel struct {
	ID       types.String                 `tfsdk:"id"`
	Hostname types.String                 `tfsdk:"hostname"`
	Routes   []resourceResourceRouteModel `tfsdk:"routes"`
	Schema   types.String                 `tfsdk:"schema"`
}

func (model *resourceDiscoveryDocumentResourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}

func (model *resourceDiscoveryDocumentResourceModel) Into() *validatorClient.ResourceSetRegistration {
	registration := validatorClient.ResourceSetRegistration{
		Hostname: model.Hostname.ValueString(),
		Routes:   make([]validatorClient.ResourceRouteRegistration, len(model.Routes)),
	}

	for i, route := range model.Routes {
		registration.Routes[i] = validatorClient.ResourceRouteRegistration{
			RouteTemplate: route.Path.ValueString(),
			ResourceUid:   route.Resource.ValueString(),
		}
	}

	return &registration
}

func (model *resourceDiscoveryDocumentResourceModel) From(source *validatorClient.ResourceSetRegistration) *resourceDiscoveryDocumentResourceModel {
	model.Hostname = types.StringValue(source.Hostname)
	model.Routes = make([]resourceResourceRouteModel, len(source.Routes))
	for i, route := range source.Routes {
		model.Routes[i] = resourceResourceRouteModel{
			Path:     types.StringValue(route.RouteTemplate),
			Resource: types.StringValue(route.ResourceUid),
		}
	}
	return model
}
