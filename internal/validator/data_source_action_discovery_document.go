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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &actionDiscoveryDocumentDataSource{}
	_ datasource.DataSourceWithConfigure = &actionDiscoveryDocumentDataSource{}
)

// NewActionDiscoveryDocumentDataSource is a helper function to simplify the provider implementation.
func NewActionDiscoveryDocumentDataSource() datasource.DataSource {
	return &actionDiscoveryDocumentDataSource{}
}

func (dataSource *actionDiscoveryDocumentDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	client := getDataSourceValidatorClient(request, response)
	if client == nil {
		// If the client is nil, we cannot proceed with the data source.
		// This method will be called again when the provider is configured,
		// so we can safely return here without setting the client.
		return
	}
	dataSource.validatorClient = client
}

// Metadata responds with the data source type name.
func (dataSource *actionDiscoveryDocumentDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_action_discovery_document"
}

// Schema defines the schema for the data source.
func (dataSource *actionDiscoveryDocumentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the action discovery document.",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "The schema that the action discovery document belongs to.",
				Required:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the action discovery document.",
				Computed:    true,
			},
			"routes": schema.ListNestedAttribute{
				Description: "The list of routes for the action discovery document.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"method": schema.StringAttribute{
							Description: "The HTTP method for the route.",
							Computed:    true,
						},
						"path": schema.StringAttribute{
							Description: "The path for the route.",
							Computed:    true,
						},
						"action": schema.StringAttribute{
							Description: "The action to be performed for the route.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (dataSource *actionDiscoveryDocumentDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading cedar schema data source")
	var configModel actionDiscoveryDocumentDataSourceModel
	err := common.ReadFromConfig(ctx, &configModel, request.Config, &response.Diagnostics)
	if err != nil {
		// If we can't read the configModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	apiData, err := dataSource.validatorClient.GetActionSet(ctx, validatorClient.GetActionSetParams{
		ID:     configModel.ID.ValueString(),
		Schema: configModel.Schema.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Resource Set", err)
		return
	}

	apiModel := &actionDiscoveryDocumentDataSourceModel{
		ID:     configModel.ID,
		Schema: configModel.Schema,
	}

	switch apiResponse := apiData.(type) {
	case *validatorClient.ActionSetRegistration:
		tflog.Debug(ctx, "Action set found, updating state")
		err = apiModel.From(apiResponse).saveToState(ctx, &response.State, &response.Diagnostics)
		if err != nil {
			common.GenerateError(&response.Diagnostics, "Saving", "Resource Set", err)
			return
		}
	case *validatorClient.GetActionSetNotFound:
		tflog.Debug(ctx, "Action set not found, setting state to empty")
		response.State.RemoveResource(ctx)
		return
	default:
		common.GenerateError(&response.Diagnostics, "Reading", "Action Set", common.ErrUnexpectedResponseType(apiResponse))
		return
	}
}

type dataSourceRouteModel struct {
	Method types.String `tfsdk:"method"`
	Path   types.String `tfsdk:"path"`
	Action types.String `tfsdk:"action"`
}

type actionDiscoveryDocumentDataSourceModel struct {
	ID       types.String           `tfsdk:"id"`
	Hostname types.String           `tfsdk:"hostname"`
	Routes   []dataSourceRouteModel `tfsdk:"routes"`
	Schema   types.String           `tfsdk:"schema"`
}

// actionDiscoveryDocumentDataSource is the data source implementation.
type actionDiscoveryDocumentDataSource struct {
	validatorClient *validatorClient.Client
}

func (model *actionDiscoveryDocumentDataSourceModel) From(source *validatorClient.ActionSetRegistration) *actionDiscoveryDocumentDataSourceModel {
	model.Hostname = types.StringValue(source.Hostname)
	model.Routes = make([]dataSourceRouteModel, len(source.Routes))
	for i, route := range source.Routes {
		model.Routes[i] = dataSourceRouteModel{
			Method: types.StringValue(route.Method),
			Path:   types.StringValue(route.RouteTemplate),
			Action: types.StringValue(route.ActionUid),
		}
	}
	return model
}

func (model *actionDiscoveryDocumentDataSourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}
