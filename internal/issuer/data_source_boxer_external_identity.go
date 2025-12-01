package issuer

import (
	"context"
	"fmt"
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
	_ datasource.DataSource              = &boxerExternalIdentityDataSource{}
	_ datasource.DataSourceWithConfigure = &boxerExternalIdentityDataSource{}
)

// NewBoxerExternalIdentityDataSource is a helper function to simplify the provider implementation.
func NewBoxerExternalIdentityDataSource() datasource.DataSource {
	return &boxerExternalIdentityDataSource{}
}

func (dataSource *boxerExternalIdentityDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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
			"validator_schema_id": schema.StringAttribute{
				Description: "The identity provider that the external identity belongs to.",
				Computed:    true,
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
	err := common.ReadFromConfig(ctx, &configModel, request.Config, &response.Diagnostics)
	if err != nil { // coverage-ignore
		// If we can't read the configModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	apiData, err := dataSource.issuerClient.GetIdentity(ctx, issuerClient.GetIdentityParams{
		ID:               configModel.ID.ValueString(),
		IdentityProvider: configModel.IdentityProvider.ValueString(),
	})
	if err != nil { // coverage-ignore
		common.GenerateError(&response.Diagnostics, "Reading", "External Identity", err)
		return
	}

	apiModel := &boxerExternalIdentityDataSourceModel{
		ID:               configModel.ID,
		IdentityProvider: configModel.IdentityProvider,
	}

	switch apiResponse := apiData.(type) {
	case *issuerClient.ExternalIdentityRegistration:
		tflog.Debug(ctx, "External identity found, updating state")
		err = apiModel.From(apiResponse).saveToState(ctx, &response.State, &response.Diagnostics)
		if err != nil { // coverage-ignore
			common.GenerateError(&response.Diagnostics, "Saving", "Resource Set", err)
			return
		}
	case *issuerClient.GetIdentityNotFound:
		tflog.Debug(ctx, "External identity not found, setting state to empty")
		response.State.RemoveResource(ctx)
		return
	default:
		common.GenerateError(&response.Diagnostics,
			"Reading",
			"External Identity",
			common.ErrUnexpectedResponseType(apiResponse))
		return
	}

}

func (model *boxerExternalIdentityDataSourceModel) From(source *issuerClient.ExternalIdentityRegistration) *boxerExternalIdentityDataSourceModel {
	model.Principal = &boxerPrincipalAssociationDataSourceModel{
		SchemaId:    types.StringValue(source.PrincipalSchema),
		PrincipalId: types.StringValue(source.PrincipalId),
	}
	model.ValidatorSchemaId = types.StringValue(source.ValidatorSchema)
	return model
}

func (model *boxerExternalIdentityDataSourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() { // coverage-ignore
		return fmt.Errorf("error getting plan")
	}
	return nil
}

type boxerPrincipalAssociationDataSourceModel struct {
	PrincipalId types.String `tfsdk:"principal_id"`
	SchemaId    types.String `tfsdk:"schema_id"`
}

type boxerExternalIdentityDataSourceModel struct {
	ID                types.String                              `tfsdk:"id"`
	IdentityProvider  types.String                              `tfsdk:"identity_provider"`
	Principal         *boxerPrincipalAssociationDataSourceModel `tfsdk:"principal"`
	ValidatorSchemaId types.String                              `tfsdk:"validator_schema_id"`
}

// boxerExternalIdentityDataSource is the data source implementation.
type boxerExternalIdentityDataSource struct {
	issuerClient *issuerClient.Client
}
