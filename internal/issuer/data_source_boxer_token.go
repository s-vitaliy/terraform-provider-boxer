package issuer

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/internal/security"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

const TokenContextKey = security.ContextKey("token")

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &boxerTokenDataSource{}
	_ datasource.DataSourceWithConfigure = &boxerTokenDataSource{}
)

// NewBoxerTokenDataSource is a helper function to simplify the provider implementation.
func NewBoxerTokenDataSource() datasource.DataSource {
	return &boxerTokenDataSource{}
}

func (dataSource *boxerTokenDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	issuerHost := getDataSourceIssuerHost(request, response)
	if issuerHost == "" {
		// If the issuerHost is nil, we cannot proceed with the data source.
		// This method will be called again when the provider is configured,
		// so we can safely return here without setting the issuerHost.
		return
	}
	client, err := issuerClient.NewClient(issuerHost, security.NewIssuerContextSecuritySource(TokenContextKey))
	if err != nil {
		response.Diagnostics.AddError(
			"Invalid Issuer Client",
			"Failed to create issuer client: "+err.Error(),
		)
		return
	}
	dataSource.issuerClient = client
}

// Metadata responds with the data source type name.
func (dataSource *boxerTokenDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_token"
}

// Schema defines the schema for the data source.
func (dataSource *boxerTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identity_provider": schema.StringAttribute{
				Description: "The identity provider that the external identity belongs to.",
				Required:    true,
			},
			"auth": schema.SingleNestedAttribute{
				Description: "The authentication details for the external identity.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"bearer": schema.StringAttribute{
						Description: "The bearer token for the external identity.",
						Required:    true,
					},
				},
			},
			"token": schema.StringAttribute{
				Description: "The token associated with the external identity.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (dataSource *boxerTokenDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var configModel boxerTokenDataSourceModel
	err := common.ReadFromConfig(ctx, &configModel, request.Config, &response.Diagnostics)
	if err != nil {
		// If we can't read the configModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	childContext := context.WithValue(ctx, TokenContextKey, configModel.Auth.Header.ValueString())
	token, err := dataSource.issuerClient.Token(childContext, issuerClient.TokenParams{
		IdentityProvider: configModel.IdentityProvider.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Getting", "Boxer token", err)
		return
	}

	tokenValue, err := io.ReadAll(token.Data)
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading to string", "Boxer token", err)
		return
	}

	model := boxerTokenDataSourceModel{
		IdentityProvider: configModel.IdentityProvider,
		Auth: &boxerTokenAuthModel{
			Header: configModel.Auth.Header,
		},
		Token: types.StringValue(string(tokenValue)),
	}

	diag := response.State.Set(ctx, &model)
	response.Diagnostics.Append(diag...)
	if response.Diagnostics.HasError() {
		return
	}
}

type boxerTokenAuthModel struct {
	Header types.String `tfsdk:"bearer"`
}

type boxerTokenDataSourceModel struct {
	IdentityProvider types.String         `tfsdk:"identity_provider"`
	Auth             *boxerTokenAuthModel `tfsdk:"auth"`
	Token            types.String         `tfsdk:"token"`
}

// boxerTokenDataSource is the data source implementation.
type boxerTokenDataSource struct {
	issuerClient *issuerClient.Client
}
