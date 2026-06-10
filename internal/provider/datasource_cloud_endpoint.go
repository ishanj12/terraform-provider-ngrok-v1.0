package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/endpoints"
)

var _ datasource.DataSource = &cloudEndpointDataSource{}

type cloudEndpointDataSourceModel struct {
	ID             types.String   `tfsdk:"id"`
	URL            types.String   `tfsdk:"url"`
	Type           types.String   `tfsdk:"type"`
	TrafficPolicy  types.String   `tfsdk:"traffic_policy"`
	Description    types.String   `tfsdk:"description"`
	Metadata       types.String   `tfsdk:"metadata"`
	Bindings       []types.String `tfsdk:"bindings"`
	PoolingEnabled types.Bool     `tfsdk:"pooling_enabled"`
	DomainID       types.String   `tfsdk:"domain_id"`
	Region         types.String   `tfsdk:"region"`
	URI            types.String   `tfsdk:"uri"`
	CreatedAt      types.String   `tfsdk:"created_at"`
	UpdatedAt      types.String   `tfsdk:"updated_at"`
}

type cloudEndpointDataSource struct {
	client *endpoints.Client
}

func NewCloudEndpointDataSource() datasource.DataSource {
	return &cloudEndpointDataSource{}
}

func (d *cloudEndpointDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_endpoint"
}

func (d *cloudEndpointDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a cloud endpoint by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique endpoint resource identifier.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL of the endpoint.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of endpoint.",
				Computed:    true,
			},
			"traffic_policy": schema.StringAttribute{
				Description: "The traffic policy attached to this endpoint.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this cloud endpoint.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this cloud endpoint.",
				Computed:    true,
			},
			"bindings": schema.ListAttribute{
				Description: "The bindings associated with this endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"pooling_enabled": schema.BoolAttribute{
				Description: "Whether the endpoint allows connection pooling.",
				Computed:    true,
			},
			"domain_id": schema.StringAttribute{
				Description: "ID of the domain reserved for this endpoint.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region of the endpoint.",
				Computed:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the cloud endpoint API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the endpoint was created, RFC 3339 format.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the endpoint was last updated, RFC 3339 format.",
				Computed:    true,
			},
		},
	}
}

func (d *cloudEndpointDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = endpoints.NewClient(clientConfig)
}

func (d *cloudEndpointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config cloudEndpointDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading cloud endpoint", err.Error())
		return
	}

	var model cloudEndpointDataSourceModel
	flattenCloudEndpointDataSource(endpoint, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenCloudEndpointDataSource(endpoint *ngrok.Endpoint, model *cloudEndpointDataSourceModel) {
	model.ID = types.StringValue(endpoint.ID)
	model.URL = types.StringValue(endpoint.URL)
	model.Type = types.StringValue(endpoint.Type)
	model.TrafficPolicy = types.StringValue(endpoint.TrafficPolicy)
	model.Description = types.StringValue(endpoint.Description)
	model.Metadata = types.StringValue(endpoint.Metadata)
	model.Bindings = flattenStringList(endpoint.Bindings)
	model.PoolingEnabled = types.BoolValue(endpoint.PoolingEnabled)
	model.DomainID = types.StringValue(flattenRef(endpoint.Domain))
	model.Region = types.StringValue(endpoint.Region)
	model.URI = types.StringValue(endpoint.URI)
	model.CreatedAt = types.StringValue(endpoint.CreatedAt)
	model.UpdatedAt = types.StringValue(endpoint.UpdatedAt)
}
