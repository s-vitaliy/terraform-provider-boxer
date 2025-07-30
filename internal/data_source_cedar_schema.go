package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-boxer/pkg/generated/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cedarSchemaDataSource{}
	_ datasource.DataSourceWithConfigure = &cedarSchemaDataSource{}
)

// NewCedarSchemaDataSource is a helper function to simplify the provider implementation.
func NewCedarSchemaDataSource() datasource.DataSource {
	return &cedarSchemaDataSource{}
}

func (dataSource *cedarSchemaDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	client := getDataSourceIssuerClient(request, response)
	if client == nil {
		// If the client is nil, we cannot proceed with the data source.
		// This method will be called again when the provider is configured,
		// so we can safely return here without setting the client.
		return
	}
	dataSource.issuerClient = client
}

// Metadata responds with the data source type name.
func (dataSource *cedarSchemaDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cedar_schema"
}

// Schema defines the schema for the data source.
func (dataSource *cedarSchemaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the cedar schema.",
				Required:    true,
			},
			"data_json": schema.StringAttribute{
				Description: "The schema data in JSON format.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (dataSource *cedarSchemaDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading cedar schema data source")
	var config cedarSchemaDataSourceModel
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	apiData, err := dataSource.issuerClient.GetSchema(ctx, issuer.GetSchemaParams{ID: config.ID.ValueString()})
	if err != nil {
		generateError(&response.Diagnostics, "Reading", "Cedar Schema", err)
		return
	}
	config.DataJson = types.StringValue(apiData.String())

	diag := response.State.Set(ctx, &config)
	response.Diagnostics.Append(diag...)
	if response.Diagnostics.HasError() {
		return
	}
}

type cedarSchemaDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	DataJson types.String `tfsdk:"data_json"`
}

// cedarSchemaDataSource is the data source implementation.
type cedarSchemaDataSource struct {
	issuerClient *issuer.Client
}
