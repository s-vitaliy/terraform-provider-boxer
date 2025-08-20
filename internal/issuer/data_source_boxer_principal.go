package issuer

import (
	"context"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &boxerPrincipalDataSource{}
	_ datasource.DataSourceWithConfigure = &boxerPrincipalDataSource{}
)

// NewBoxerPrincipalDataSource is a helper function to simplify the provider implementation.
func NewBoxerPrincipalDataSource() datasource.DataSource {
	return &boxerPrincipalDataSource{}
}

func (dataSource *boxerPrincipalDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func (dataSource *boxerPrincipalDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_principal"
}

// Schema defines the schema for the data source.
func (dataSource *boxerPrincipalDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the cedar schema.",
				Required:    true,
			},
			"schema_id": schema.StringAttribute{
				Description: "The schema ID that this principal belongs to.",
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
func (dataSource *boxerPrincipalDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var configModel boxerPrincipalDataSourceModel
	err := common.ReadFromConfig(ctx, &configModel, request.Config, &response.Diagnostics)
	if err != nil {
		// If we can't read the configModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	params := issuerClient.GetPrincipalParams{
		ID:     configModel.ID.ValueString(),
		Schema: configModel.SchemaId.ValueString(),
	}

	apiData, err := dataSource.issuerClient.GetPrincipal(ctx, params)
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Boxer Principal", err)
		return
	}

	switch apiResponse := apiData.(type) {
	case *issuerClient.GetPrincipalOKApplicationJSON:
		configModel.DataJson = types.StringValue(jx.Raw(*apiResponse).String())
		diag := response.State.Set(ctx, &configModel)
		response.Diagnostics.Append(diag...)
		if response.Diagnostics.HasError() {
			return
		}
		return
	case *issuerClient.GetPrincipalNotFound:
		response.State.RemoveResource(ctx)
		return
	default:
		common.GenerateError(&response.Diagnostics,
			"Reading",
			"Boxer Principal",
			common.ErrUnexpectedResponseType(apiData))
		return
	}
}

type boxerPrincipalDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	DataJson types.String `tfsdk:"data_json"`
	SchemaId types.String `tfsdk:"schema_id"`
}

// boxerPrincipalDataSource is the data source implementation.
type boxerPrincipalDataSource struct {
	issuerClient *issuerClient.Client
}
