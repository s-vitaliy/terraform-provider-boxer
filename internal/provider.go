package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func (b BoxerProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "boxer"
	response.Version = b.version
}

func (b BoxerProvider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{}
}

func (b BoxerProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
}

func (b BoxerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (b BoxerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BoxerProvider{
			version: version,
		}
	}
}
