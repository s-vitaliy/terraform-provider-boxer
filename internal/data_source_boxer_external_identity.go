package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-boxer/pkg/generated/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &boxerExternalIdentityDataSource{}
	_ datasource.DataSourceWithConfigure = &boxerExternalIdentityDataSource{}
)

// NewBoxerExternalIdentityDataSource is a helper function to simplify the provider implementation.
func NewBoxerExternalIdentityDataSource() datasource.DataSource {
	return &boxerExternalIdentityDataSource{}
}

func (dataSource *boxerExternalIdentityDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
func (dataSource *boxerExternalIdentityDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_external_identity"
}

// Schema defines the schema for the data source.
func (dataSource *boxerExternalIdentityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the principal.",
				Required:    true,
			},
			"identity_provider": schema.StringAttribute{
				Description: "The identity provider that the external identity belongs to.",
				Required:    true,
			},

			"principal": schema.SingleNestedAttribute{
				Description: "The principal ID associated to the external identity.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"principal_id": schema.StringAttribute{
						Description: "The unique identifier of the principal associated with the external identity.",
						Computed:    true,
					},
					"schema_id": schema.StringAttribute{
						Description: "The schema ID of the principal associated with the external identity.",
						Computed:    true,
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (dataSource *boxerExternalIdentityDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var configModel boxerExternalIdentityDataSourceModel
	err := readFromConfig(ctx, &configModel, request.Config, &response.Diagnostics)
	if err != nil {
		// If we can't read the configModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	externalIdentity, err := dataSource.issuerClient.GetIdentity(ctx, issuer.GetIdentityParams{
		ID:               configModel.ID.ValueString(),
		IdentityProvider: configModel.IdentityProvider.ValueString(),
	})
	if err != nil {
		generateError(&response.Diagnostics, "Reading", "External Identity", err)
		return
	}

	externalIdentityAssociation, err := dataSource.issuerClient.GetAssociation(ctx, issuer.GetAssociationParams{
		ID:               configModel.ID.ValueString(),
		IdentityProvider: configModel.IdentityProvider.ValueString(),
	})
	if err != nil {
		generateError(&response.Diagnostics, "Reading", "External Identity association", err)
		return
	}
	model := boxerExternalIdentityModel{
		ID:               types.StringValue(externalIdentity.UserId),
		IdentityProvider: types.StringValue(externalIdentity.IdentityProvider),
		Principal: boxerPrincipalAssociationModel{
			PrincipalId: types.StringValue(externalIdentityAssociation.PrincipalID),
			SchemaId:    types.StringValue(externalIdentityAssociation.PrincipalSchema),
		},
	}

	diag := response.State.Set(ctx, &model)
	response.Diagnostics.Append(diag...)
	if response.Diagnostics.HasError() {
		return
	}
}

type boxerExternalIdentityDataSourceModel struct {
	ID               types.String                    `tfsdk:"id"`
	IdentityProvider types.String                    `tfsdk:"identity_provider"`
	Principal        *boxerPrincipalAssociationModel `tfsdk:"principal"`
}

// boxerExternalIdentityDataSource is the data source implementation.
type boxerExternalIdentityDataSource struct {
	issuerClient *issuer.Client
}
