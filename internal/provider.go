package internal

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"terraform-provider-boxer/internal/issuer"
	"terraform-provider-boxer/internal/security"
	"terraform-provider-boxer/internal/validator"
	"terraform-provider-boxer/pkg/generated/api/issuerClient"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &BoxerProvider{}
)

// BoxerProvider struct implements the Boxer Terraform provider
type BoxerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (b BoxerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "boxer"
	response.Version = b.version
}

func (b BoxerProvider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"issuer_host": schema.StringAttribute{
				Description: "The host of the Boxer Issuer API.",
				Required:    true,
			},
			"validator_host": schema.StringAttribute{
				Description: "The host of the Boxer Issuer API.",
				Required:    true,
			},
			"external_auth": schema.SingleNestedAttribute{
				Description: "Configuration for external authentication",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"security_token": schema.StringAttribute{
						Description: "An external security token to use for Boxer Issuer API calls. " +
							"If not set, the BOXER_EXTERNAL_SECURITY_TOKEN environment variable will be used if set.",
						Optional:  true,
						Sensitive: true,
					},
					"identity_provider_id": schema.StringAttribute{
						Description: "The ID of the external identity provider to use for authentication.",
						Required:    true,
						Sensitive:   false,
					},
					"internal_token_provider_endpoint": schema.StringAttribute{
						Description: "The token endpoint of the external identity provider.",
						Required:    true,
					},
				},
			},
		},
	}
}

type externalAuthModel struct {
	SecurityToken                 types.String `tfsdk:"security_token"`
	IdentityProviderID            types.String `tfsdk:"identity_provider_id"`
	InternalTokenProviderEndpoint types.String `tfsdk:"internal_token_provider_endpoint"`
}

type boxerProviderModel struct {
	IssuerHost    types.String      `tfsdk:"issuer_host"`
	ValidatorHost types.String      `tfsdk:"validator_host"`
	ExternalAuth  externalAuthModel `tfsdk:"external_auth"`
}

func (b BoxerProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config boxerProviderModel
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() { // coverage-ignore
		return
	}
	if config.IssuerHost.IsUnknown() { // coverage-ignore
		response.Diagnostics.AddAttributeError(
			path.Root("issuer_host"),
			"Unknown Boxer Issuer issuerHost",
			"The issuer_host parameter is mandatory"+
				"Either target apply the source of the value first, set the value statically in the configuration,"+
				"or use the BOXER_ISSUER_HOST environment variable.",
		)
	}

	if response.Diagnostics.HasError() { // coverage-ignore
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	issuerHost := os.Getenv("BOXER_ISSUER_HOST")
	validatorHost := os.Getenv("BOXER_VALIDATOR_HOST")

	if !config.IssuerHost.IsNull() {
		issuerHost = config.IssuerHost.ValueString()
	}

	if !config.ValidatorHost.IsNull() {
		validatorHost = config.ValidatorHost.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if issuerHost == "" { // coverage-ignore
		response.Diagnostics.AddAttributeError(
			path.Root("issuer_host"),
			"Unknown Boxer Issuer issuerHost",
			"The issuer_host parameter is mandatory"+
				"Either target apply the source of the value first, set the value statically in the configuration,"+
				"or use the BOXER_ISSUER_HOST environment variable.",
		)
	}

	if response.Diagnostics.HasError() { // coverage-ignore
		return
	}

	if config.ExternalAuth.SecurityToken.IsNull() {
		token, ok := os.LookupEnv("BOXER_EXTERNAL_SECURITY_TOKEN")
		if !ok { // coverage-ignore
			response.Diagnostics.AddAttributeError(
				path.Root("external_security_token"),
				"Missing Boxer Issuer externalSecurityToken",
				"The external_security_token parameter is required if the BOXER_EXTERNAL_SECURITY_TOKEN environment variable is not set."+
					"Either set the value statically in the configuration, or use the BOXER_EXTERNAL_SECURITY_TOKEN environment variable.",
			)
			return
		}
		config.ExternalAuth.SecurityToken = types.StringValue(token)
	}

	tokenEndpoint := config.ExternalAuth.InternalTokenProviderEndpoint.ValueString()
	externalSecuritySource := security.IssuerStaticSecuritySource(config.ExternalAuth.SecurityToken.ValueString())
	externalAuthIssuerClient, err := issuerClient.NewClient(tokenEndpoint, externalSecuritySource)
	if err != nil { // coverage-ignore
		response.Diagnostics.AddError(
			"Failed to initialize Self-authorization client",
			"An unexpected error occurred when creating the Boxer Issuer client: "+err.Error(),
		)
		return
	}

	internalTokenReader := security.NewInternalTokenReader(externalAuthIssuerClient, config.ExternalAuth.IdentityProviderID.ValueString())

	issuerApiClient, err := issuerClient.NewClient(issuerHost, security.NewIssuerInternalSecuritySource(internalTokenReader))
	if err != nil { // coverage-ignore
		response.Diagnostics.AddError(
			"Failed to initialize Boxer Issuer Client",
			"An unexpected error occurred when creating the Boxer Issuer client: "+err.Error(),
		)
		return
	}

	validatorApiClient, err := validatorClient.NewClient(validatorHost, security.NewValidatorInternalSecuritySource(internalTokenReader))
	if err != nil { // coverage-ignore
		response.Diagnostics.AddError(
			"Failed to initialize Boxer Validator Client",
			"An unexpected error occurred when creating the Boxer Validator client: "+err.Error(),
		)
		return
	}

	data := &BoxerProviderData{
		issuerClient:    issuerApiClient,
		validatorClient: validatorApiClient,
		issuerHost:      issuerHost,
	}
	response.DataSourceData = data
	response.ResourceData = data
}

func (b BoxerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Issuer Data sources
		issuer.NewIdentityProviderDataSource,
		issuer.NewCedarSchemaDataSource,
		issuer.NewBoxerPrincipalDataSource,
		issuer.NewBoxerExternalIdentityDataSource,
		issuer.NewBoxerTokenDataSource,

		// Validator data sources
		validator.NewCedarSchemaDataSource,
		validator.NewActionDiscoveryDocumentDataSource,
		validator.NewResourceDiscoveryDocumentDataSource,
		validator.NewPolicySetDataSource,
	}
}

func (b BoxerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Boxer Issuer resources
		issuer.NewIdentityProviderResource,
		issuer.NewCedarSchemaResource,
		issuer.NewBoxerPrincipalResource,
		issuer.NewBoxerExternalIdentityResource,

		// Boxer Validator resources
		validator.NewCedarSchemaResource,
		validator.NewActionDiscoveryDocumentResource,
		validator.NewResourceDiscoveryDocumentResource,
		validator.NewPolicySetResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BoxerProvider{
			version: version,
		}
	}
}
