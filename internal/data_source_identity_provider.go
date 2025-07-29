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
	_ datasource.DataSource              = &identityProviderDataSource{}
	_ datasource.DataSourceWithConfigure = &identityProviderDataSource{}
)

// NewIdentityProviderDataSource is a helper function to simplify the provider implementation.
func NewIdentityProviderDataSource() datasource.DataSource {
	return &identityProviderDataSource{}
}

func (d *identityProviderDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
	d.issuerClient = data.issuerClient
}

// Metadata returns the data source type name.
func (d *identityProviderDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_identity_provider"
}

// Schema defines the schema for the data source.
func (d *identityProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the identity provider.",
				Required:    true,
			},
			"discovery_url": schema.StringAttribute{
				Description: "The OIDC discovery URL of the identity provider.",
				Computed:    true,
			},
			"user_id_claim": schema.StringAttribute{
				Description: "The claim used to identify the user in the identity provider's token.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *identityProviderDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading identity provider data source")
	var id identityProviderDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &id)...)
	apiData, err := d.issuerClient.GetProvider(ctx, issuer.GetProviderParams{ID: id.ID.ValueString()})
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading identity provider",
			"An error occurred while reading the identity provider: "+err.Error(),
		)
		return
	}
	id.DiscoveryUrl = types.StringValue(apiData.GetDiscoveryUrl())
	id.UserIdClaim = types.StringValue(apiData.GetUserIdClaim())

	diag := response.State.Set(ctx, &id)
	response.Diagnostics.Append(diag...)
}

type identityProviderDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	DiscoveryUrl types.String `tfsdk:"discovery_url"`
	UserIdClaim  types.String `tfsdk:"user_id_claim"`
	//Audiences    types.List   `tfsdk:"audiences"`
	//Issuers      types.List   `json:"issuers"`
}

// identityProviderDataSource is the data source implementation.
type identityProviderDataSource struct {
	issuerClient *issuer.Client
}
