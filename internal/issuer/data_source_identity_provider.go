package issuer

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
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

func (dataSource *identityProviderDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	client := getDataSourceIssuerClient(request, response)
	if client == nil { // coverage-ignore
		// If the client is nil, we cannot proceed with the data source.
		// This method will be called again when the provider is configured,
		// so we can safely return here without setting the client.
		return
	}
	dataSource.issuerClient = client
}

// Metadata responds with the data source type name.
func (dataSource *identityProviderDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_identity_provider"
}

// Schema defines the schema for the data source.
func (dataSource *identityProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			"audiences": schema.ListAttribute{
				Description: "The list of audiences accepted by the identity provider.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"issuers": schema.ListAttribute{
				Description: "The list of issuers accepted by the identity provider.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (dataSource *identityProviderDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading identity provider data source")
	var configModel identityProviderDataSourceModel
	err := common.ReadFromConfig(ctx, &configModel, request.Config, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the config, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}

	apiData, err := dataSource.issuerClient.GetProvider(ctx, issuerClient.GetProviderParams{ID: configModel.ID.ValueString()})
	if err != nil { // coverage-ignore
		common.GenerateError(&response.Diagnostics, "Reading", "Identity Provider", err)
		return
	}

	switch apiResponse := apiData.(type) {
	case *issuerClient.OidcIdentityProviderRegistration:
		tflog.Debug(ctx, "Identity provider found, updating state")
		err := configModel.From(apiResponse).saveToState(ctx, &response.State, &response.Diagnostics)
		if err != nil { // coverage-ignore
			return
		}
	case *issuerClient.GetProviderNotFound:
		tflog.Debug(ctx, "Identity provider not found, setting state to empty")
		response.State.RemoveResource(ctx)
		return
	default:
		common.GenerateError(&response.Diagnostics, "Reading", "Identity Provider", common.ErrUnexpectedResponseType(apiResponse))
		return
	}
}

type identityProviderDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	DiscoveryUrl types.String `tfsdk:"discovery_url"`
	UserIdClaim  types.String `tfsdk:"user_id_claim"`
	Audiences    types.List   `tfsdk:"audiences"`
	Issuers      types.List   `tfsdk:"issuers"`
}

func (model *identityProviderDataSourceModel) From(source *issuerClient.OidcIdentityProviderRegistration) *identityProviderDataSourceModel {
	model.DiscoveryUrl = types.StringValue(source.GetDiscoveryUrl())
	model.UserIdClaim = types.StringValue(source.GetUserIdClaim())
	issuers := make([]attr.Value, len(source.GetIssuers()))
	for i, issuer := range source.GetIssuers() {
		issuers[i] = types.StringValue(issuer)
	}
	model.Issuers = types.ListValueMust(types.StringType, issuers)

	audiences := make([]attr.Value, len(source.GetAudiences()))
	for i, audience := range source.GetAudiences() {
		audiences[i] = types.StringValue(audience)
	}
	model.Audiences = types.ListValueMust(types.StringType, audiences)
	return model
}

func (model *identityProviderDataSourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error saving state")
	}
	return nil
}

// identityProviderDataSource is the data source implementation.
type identityProviderDataSource struct {
	issuerClient *issuerClient.Client
}
