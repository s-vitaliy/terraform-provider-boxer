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
		},
	}
}

type boxerProviderModel struct {
	IssuerHost    types.String `tfsdk:"issuer_host"`
	ValidatorHost types.String `tfsdk:"validator_host"`
}

func (b BoxerProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config boxerProviderModel
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if config.IssuerHost.IsUnknown() {
		response.Diagnostics.AddAttributeError(
			path.Root("issuer_host"),
			"Unknown Boxer Issuer issuerHost",
			"The issuer_host parameter is mandatory"+
				"Either target apply the source of the value first, set the value statically in the configuration,"+
				"or use the BOXER_ISSUER_HOST environment variable.",
		)
	}

	if response.Diagnostics.HasError() {
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

	if issuerHost == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("issuer_host"),
			"Unknown Boxer Issuer issuerHost",
			"The issuer_host parameter is mandatory"+
				"Either target apply the source of the value first, set the value statically in the configuration,"+
				"or use the BOXER_ISSUER_HOST environment variable.",
		)
	}

	if response.Diagnostics.HasError() {
		return
	}

	issuerApiClient, err := issuerClient.NewClient(issuerHost, security.NewEmptySecuritySource())
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to initialize Boxer Issuer Client",
			"An unexpected error occurred when creating the Boxer Issuer client: "+err.Error(),
		)
		return
	}

	validatorApiClient, err := validatorClient.NewClient(validatorHost)
	if err != nil {
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
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BoxerProvider{
			version: version,
		}
	}
}
