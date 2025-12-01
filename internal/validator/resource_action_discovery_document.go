package validator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	_ resource.Resource              = &actionDiscoveryDocumentResource{}
	_ resource.ResourceWithConfigure = &actionDiscoveryDocumentResource{}
)

// NewActionDiscoveryDocumentResource is a helper function to simplify the provider implementation.
func NewActionDiscoveryDocumentResource() resource.Resource {
	return &actionDiscoveryDocumentResource{}
}

// actionDiscoveryDocumentResource is the resource implementation.
type actionDiscoveryDocumentResource struct {
	validatorClient *validatorClient.Client
}

func (resource *actionDiscoveryDocumentResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := getResourceValidatorClient(request, response)
	resource.validatorClient = client
}

// Metadata responds with the resource type name.
func (resource *actionDiscoveryDocumentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_action_discovery_document"
}

// Schema defines the schema for the resource.
func (resource *actionDiscoveryDocumentResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the action discovery document.",
				Required:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the action discovery document.",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "The schema that the action discovery document belongs to.",
				Required:    true,
			},
			"routes": schema.ListNestedAttribute{
				Description: "The list of routes for the action discovery document.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"method": schema.StringAttribute{
							Description: "The HTTP method for the route.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"),
							},
						},
						"path": schema.StringAttribute{
							Description: "The path for the route.",
							Required:    true,
						},
						"action": schema.StringAttribute{
							Description: "The action to be performed for the route.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (resource *actionDiscoveryDocumentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var planModel actionDiscoveryDocumentResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.validatorClient.PostActionSet(ctx, planModel.Into(), validatorClient.PostActionSetParams{
		ID:     planModel.ID.ValueString(),
		Schema: planModel.Schema.ValueString(),
	})

	if err != nil { // coverage-ignore
		common.GenerateError(&response.Diagnostics, "Creating", "Resource Set", err)
		return
	}

	err = planModel.saveToState(ctx, &response.State, &response.Diagnostics)
	// If we can't save the state, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil { // coverage-ignore
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (resource *actionDiscoveryDocumentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var stateModel actionDiscoveryDocumentResourceModel
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
	apiData, err := resource.validatorClient.GetActionSet(ctx, validatorClient.GetActionSetParams{
		ID:     stateModel.ID.ValueString(),
		Schema: stateModel.Schema.ValueString(),
	})

	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Resource Set", err)
		return
	}

	apiModel := &actionDiscoveryDocumentResourceModel{
		ID:     stateModel.ID,
		Schema: stateModel.Schema,
	}

	switch apiResponse := apiData.(type) {
	case *validatorClient.ActionSetRegistration:
		err = apiModel.From(apiResponse).saveToState(ctx, &response.State, &response.Diagnostics)
		// If we can't save the stateModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		if err != nil { // coverage-ignore
			return
		}
	case *validatorClient.GetActionSetNotFound:
		tflog.Debug(ctx, "Action set not found, setting state to empty")
		response.State.RemoveResource(ctx)
		return
	default:
		common.GenerateError(&response.Diagnostics,
			"Reading",
			"Action Set",
			common.ErrUnexpectedResponseType(apiResponse))
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (resource *actionDiscoveryDocumentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var planModel actionDiscoveryDocumentResourceModel
	err := common.ReadFromPlan(ctx, &planModel, request.Plan, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the planModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.validatorClient.PostActionSet(ctx, planModel.Into(), validatorClient.PostActionSetParams{
		ID:     planModel.ID.ValueString(),
		Schema: planModel.Schema.ValueString(),
	})
	if err != nil { // coverage-ignore
		common.GenerateError(&response.Diagnostics, "Updating", "Resource Set", err)
		return
	}

	err = planModel.saveToState(ctx, &response.State, &response.Diagnostics)
	// If we can't save the stateModel, we can't proceed with the update.
	// so we return early.
	// The error will be handled by the framework and returned to the user.
	if err != nil { // coverage-ignore
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (resource *actionDiscoveryDocumentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var stateModel actionDiscoveryDocumentResourceModel
	err := common.ReadFromState(ctx, &stateModel, request.State, &response.Diagnostics)

	if err != nil { // coverage-ignore
		// If we can't read the state, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	err = resource.validatorClient.DeleteActionSet(ctx, validatorClient.DeleteActionSetParams{
		ID:     stateModel.ID.ValueString(),
		Schema: stateModel.Schema.ValueString(),
	})

	if err != nil { // coverage-ignore
		common.GenerateError(&response.Diagnostics, "Deleting", "Resource Set", err)
		return
	}
}

type resourceActionRouteModel struct {
	Method types.String `tfsdk:"method"`
	Path   types.String `tfsdk:"path"`
	Action types.String `tfsdk:"action"`
}

type actionDiscoveryDocumentResourceModel struct {
	ID       types.String               `tfsdk:"id"`
	Hostname types.String               `tfsdk:"hostname"`
	Routes   []resourceActionRouteModel `tfsdk:"routes"`
	Schema   types.String               `tfsdk:"schema"`
}

func (model *actionDiscoveryDocumentResourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() { // coverage-ignore
		return fmt.Errorf("error getting plan")
	}
	return nil
}

func (model *actionDiscoveryDocumentResourceModel) Into() *validatorClient.ActionSetRegistration {
	registration := validatorClient.ActionSetRegistration{
		Hostname: model.Hostname.ValueString(),
		Routes:   make([]validatorClient.ActionRouteRegistration, len(model.Routes)),
	}

	for i, route := range model.Routes {
		registration.Routes[i] = validatorClient.ActionRouteRegistration{
			Method:        route.Method.ValueString(),
			RouteTemplate: route.Path.ValueString(),
			ActionUid:     route.Action.ValueString(),
		}
	}

	return &registration
}

func (model *actionDiscoveryDocumentResourceModel) From(source *validatorClient.ActionSetRegistration) *actionDiscoveryDocumentResourceModel {
	model.Hostname = types.StringValue(source.Hostname)
	model.Routes = make([]resourceActionRouteModel, len(source.Routes))
	for i, route := range source.Routes {
		model.Routes[i] = resourceActionRouteModel{
			Method: types.StringValue(route.Method),
			Path:   types.StringValue(route.RouteTemplate),
			Action: types.StringValue(route.ActionUid),
		}
	}
	return model
}
