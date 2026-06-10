package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/application_sessions"
)

var _ datasource.DataSource = &applicationSessionDataSource{}

type applicationSessionDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	URI               types.String `tfsdk:"uri"`
	PublicURL         types.String `tfsdk:"public_url"`
	ApplicationUserID types.String `tfsdk:"application_user_id"`
	CreatedAt         types.String `tfsdk:"created_at"`
	LastActive        types.String `tfsdk:"last_active"`
	ExpiresAt         types.String `tfsdk:"expires_at"`
	EndpointID        types.String `tfsdk:"endpoint_id"`
	EdgeID            types.String `tfsdk:"edge_id"`
	RouteID           types.String `tfsdk:"route_id"`
}

type applicationSessionDataSource struct {
	client *application_sessions.Client
}

func NewApplicationSessionDataSource() datasource.DataSource {
	return &applicationSessionDataSource{}
}

func (d *applicationSessionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_session"
}

func (d *applicationSessionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up an application session by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique application session resource identifier.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the application session API resource.",
				Computed:    true,
			},
			"public_url": schema.StringAttribute{
				Description: "The public URL of the application session.",
				Computed:    true,
			},
			"application_user_id": schema.StringAttribute{
				Description: "The ID of the application user associated with this session.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the application session was created, in RFC 3339 format.",
				Computed:    true,
			},
			"last_active": schema.StringAttribute{
				Description: "Timestamp when the application session was last active, in RFC 3339 format.",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "Timestamp when the application session expires, in RFC 3339 format.",
				Computed:    true,
			},
			"endpoint_id": schema.StringAttribute{
				Description: "The ID of the endpoint associated with this session.",
				Computed:    true,
			},
			"edge_id": schema.StringAttribute{
				Description: "The ID of the edge associated with this session.",
				Computed:    true,
			},
			"route_id": schema.StringAttribute{
				Description: "The ID of the route associated with this session.",
				Computed:    true,
			},
		},
	}
}

func (d *applicationSessionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = application_sessions.NewClient(clientConfig)
}

func (d *applicationSessionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config applicationSessionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	session, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading application session", err.Error())
		return
	}

	var model applicationSessionDataSourceModel
	flattenApplicationSessionDataSource(session, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenApplicationSessionDataSource(s *ngrok.ApplicationSession, model *applicationSessionDataSourceModel) {
	model.ID = types.StringValue(s.ID)
	model.URI = types.StringValue(s.URI)
	model.PublicURL = types.StringValue(s.PublicURL)
	model.ApplicationUserID = types.StringValue(flattenRef(s.ApplicationUser))
	model.CreatedAt = types.StringValue(s.CreatedAt)
	model.LastActive = types.StringValue(s.LastActive)
	model.ExpiresAt = types.StringValue(s.ExpiresAt)
	model.EndpointID = types.StringValue(flattenRef(s.Endpoint))
	model.EdgeID = types.StringValue(flattenRef(s.Edge))
	model.RouteID = types.StringValue(flattenRef(s.Route))
}
