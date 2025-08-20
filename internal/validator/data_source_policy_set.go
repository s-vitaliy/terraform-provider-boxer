package validator

import (
	"context"
	"fmt"
	"github.com/cedar-policy/cedar-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-boxer/internal/common"
	"terraform-provider-boxer/pkg/generated/api/validatorClient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &policySetDataSource{}
	_ datasource.DataSourceWithConfigure = &policySetDataSource{}
)

// NewPolicySetDataSource is a helper function to simplify the provider implementation.
func NewPolicySetDataSource() datasource.DataSource {
	return &policySetDataSource{}
}

func (dataSource *policySetDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	client := getDataSourceValidatorClient(request, response)
	if client == nil {
		// If the client is nil, we cannot proceed with the data source.
		// This method will be called again when the provider is configured,
		// so we can safely return here without setting the client.
		return
	}
	dataSource.validatorClient = client
}

// Metadata responds with the data source type name.
func (dataSource *policySetDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_policy_set"
}

// Schema defines the schema for the data source.
func (dataSource *policySetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the policy set.",
				Required:    true,
			},
			"schema": schema.StringAttribute{
				Description: "The schema that the policy set conforms to.",
				Required:    true,
			},
			"data_cedar": schema.StringAttribute{
				Description: "The Cedar schema data in Cedar format.",
				Computed:    true,
			},
			"data_json": schema.StringAttribute{
				Description: "The Cedar schema data in Cedar format.",
				Computed:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (dataSource *policySetDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var configModel policySetDataSourceModel
	err := common.ReadFromConfig(ctx, &configModel, request.Config, &response.Diagnostics)
	if err != nil {
		// If we can't read the configModel, we can't proceed with the update.
		// so we return early.
		// The error will be handled by the framework and returned to the user.
		return
	}
	apiData, err := dataSource.validatorClient.GetPolicySet(ctx, validatorClient.GetPolicySetParams{
		ID:     configModel.ID.ValueString(),
		Schema: configModel.Schema.ValueString(),
	})
	if err != nil {
		common.GenerateError(&response.Diagnostics, "Reading", "Resource Set", err)
		return
	}

	apiModel := &policySetDataSourceModel{
		ID:     configModel.ID,
		Schema: configModel.Schema,
	}
	switch apiResponse := apiData.(type) {
	case *validatorClient.PolicySetRegistration:
		apiModel, err = apiModel.From(apiResponse)
		if err != nil {
			common.GenerateError(&response.Diagnostics, "Converting", "Resource Set", err)
			return
		}

		err = apiModel.saveToState(ctx, &response.State, &response.Diagnostics)
		if err != nil {
			common.GenerateError(&response.Diagnostics, "Saving", "Resource Set", err)
			return
		}
		return
	case *validatorClient.GetPolicySetNotFound:
		tflog.Debug(ctx, "Policy set not found, setting state to empty")
		response.State.RemoveResource(ctx)
		return
	default:
		common.GenerateError(&response.Diagnostics, "Reading", "Policy Set", common.ErrUnexpectedResponseType(apiResponse))
		return
	}

}

type policySetDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	DataCedar types.String `tfsdk:"data_cedar"`
	DataJson  types.String `tfsdk:"data_json"`
	Schema    types.String `tfsdk:"schema"`
}

// policySetDataSource is the data source implementation.
type policySetDataSource struct {
	validatorClient *validatorClient.Client
}

func (model *policySetDataSourceModel) From(source *validatorClient.PolicySetRegistration) (*policySetDataSourceModel, error) {
	var policy cedar.Policy
	err := policy.UnmarshalCedar([]byte(source.Policy))
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal Cedar policy: %v", err))
	}

	valueJson, err := policy.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshalling Cedar policy to JSON: %w", err)
	}

	model.DataJson = types.StringValue(string(valueJson))
	model.DataCedar = types.StringValue(string(policy.MarshalCedar()))

	return model, nil
}

func (model *policySetDataSourceModel) saveToState(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics) error {
	diags := state.Set(ctx, &model)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return fmt.Errorf("error getting plan")
	}
	return nil
}
