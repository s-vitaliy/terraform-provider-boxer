package validator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &resourceDiscoveryDocumentDataSource{}
	_ datasource.DataSourceWithConfigure = &resourceDiscoveryDocumentDataSource{}
)

// NewResourceDiscoveryDocumentDataSource is a helper function to simplify the provider implementation.
func NewResourceDiscoveryDocumentDataSource() datasource.DataSource {
	return &resourceDiscoveryDocumentDataSource{}
}

func (dataSource *resourceDiscoveryDocumentDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func (dataSource *resourceDiscoveryDocumentDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_resource_discovery_document"
}

// Schema defines the schema for the data source.
func (dataSource *resourceDiscoveryDocumentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the resource discovery document.",
				Required:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the resource discovery document.",
				Computed:    true,
			},
			"schema": schema.StringAttribute{
				Description: "The schema that the action discovery document belongs to.",
				Computed:    true, // TODO: fix here and in other files
			},
			"routes": schema.ListNestedAttribute{
				Description: "The list of routes for the resource discovery document.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Description: "The path for the route.",
							Computed:    true,
						},
						"resource": schema.StringAttribute{
							Description: "The resource associated with the route.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (dataSource *resourceDiscoveryDocumentDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var configModel resourceDiscoveryDocumentDataSourceModel
	err := common.ReadFromConfig(ctx, &configModel, request.Config, &response.Diagnostics)
	if err != nil {
		// If we can't read the configModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	registration, err := dataSource.validatorClient.GetResourceSet(ctx, validatorClient.GetResourceSetParams{
		ID:     configModel.ID.ValueString(),
		Schema: configModel.Schema.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Resource Set", err)
		return
	}

	apiModel := &resourceDiscoveryDocumentDataSourceModel{
		ID:     configModel.ID,
		Schema: configModel.Schema,
	}

	err = apiModel.From(registration).saveToState(ctx, &response.State, &response.Diagnostics)

	if err != nil {
		common.GenerateError(&response.Diagnostics, "Saving", "Resource Set", err)
		return
	}
}

type resourceDataSourceRouteModel struct {
	Path     types.String `tfsdk:"path"`
	Resource types.String `tfsdk:"resource"`
}

type resourceDiscoveryDocumentDataSourceModel struct {
	ID       types.String                   `tfsdk:"id"`
	Hostname types.String                   `tfsdk:"hostname"`
	Routes   []resourceDataSourceRouteModel `tfsdk:"routes"`
	Schema   types.String                   `tfsdk:"schema"`
}

// resourceDiscoveryDocumentDataSource is the data source implementation.
type resourceDiscoveryDocumentDataSource struct {
	validatorClient *validatorClient.Client
}

func (model *resourceDiscoveryDocumentDataSourceModel) From(source *validatorClient.ResourceSetRegistration) *resourceDiscoveryDocumentDataSourceModel {
	model.Hostname = types.StringValue(source.Hostname)
	model.Routes = make([]resourceDataSourceRouteModel, len(source.Routes))
	for i, route := range source.Routes {
		model.Routes[i] = resourceDataSourceRouteModel{
			Path:     types.StringValue(route.RouteTemplate),
			Resource: types.StringValue(route.ResourceUid),
		}
	}
	return model
}

func (model *resourceDiscoveryDocumentDataSourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}
