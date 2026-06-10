package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/api_keys"
)

var _ datasource.DataSource = &apiKeyDataSource{}

type apiKeyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	URI         types.String `tfsdk:"uri"`
	Description types.String `tfsdk:"description"`
	Metadata    types.String `tfsdk:"metadata"`
	CreatedAt   types.String `tfsdk:"created_at"`
	OwnerID     types.String `tfsdk:"owner_id"`
}

type apiKeyDataSource struct {
	client *api_keys.Client
}

func NewAPIKeyDataSource() datasource.DataSource {
	return &apiKeyDataSource{}
}

func (d *apiKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *apiKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up an API key by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique API key resource identifier.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the API key API resource.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of what this API key is used for.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this API key.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the API key was created, RFC 3339 format.",
				Computed:    true,
			},
			"owner_id": schema.StringAttribute{
				Description: "The owner ID of the API key.",
				Computed:    true,
			},
		},
	}
}

func (d *apiKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientConfig, ok := req.ProviderData.(*ngrok.ClientConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ngrok.ClientConfig, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = api_keys.NewClient(clientConfig)
}

func (d *apiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config apiKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading API key", err.Error())
		return
	}

	var model apiKeyDataSourceModel
	model.ID = types.StringValue(apiKey.ID)
	model.URI = types.StringValue(apiKey.URI)
	model.Description = types.StringValue(apiKey.Description)
	model.Metadata = types.StringValue(apiKey.Metadata)
	model.CreatedAt = types.StringValue(apiKey.CreatedAt)

	if apiKey.OwnerID != nil {
		model.OwnerID = types.StringValue(*apiKey.OwnerID)
	} else {
		model.OwnerID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
