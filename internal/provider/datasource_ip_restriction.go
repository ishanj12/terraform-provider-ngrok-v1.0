package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/ip_restrictions"
)

var _ datasource.DataSource = &ipRestrictionDataSource{}

type ipRestrictionDataSourceModel struct {
	ID          types.String   `tfsdk:"id"`
	URI         types.String   `tfsdk:"uri"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	Description types.String   `tfsdk:"description"`
	Metadata    types.String   `tfsdk:"metadata"`
	Enforced    types.Bool     `tfsdk:"enforced"`
	Type        types.String   `tfsdk:"type"`
	IPPolicyIDs []types.String `tfsdk:"ip_policy_ids"`
}

type ipRestrictionDataSource struct {
	client *ip_restrictions.Client
}

func NewIPRestrictionDataSource() datasource.DataSource {
	return &ipRestrictionDataSource{}
}

func (d *ipRestrictionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_restriction"
}

func (d *ipRestrictionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up an IP restriction by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this IP restriction.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the IP restriction API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the IP restriction was created, RFC 3339 format.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this IP restriction.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this IP restriction.",
				Computed:    true,
			},
			"enforced": schema.BoolAttribute{
				Description: "True if the IP restriction will be enforced.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the IP restriction.",
				Computed:    true,
			},
			"ip_policy_ids": schema.ListAttribute{
				Description: "The set of IP policy identifiers that are used to enforce the restriction.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *ipRestrictionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = ip_restrictions.NewClient(clientConfig)
}

func (d *ipRestrictionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ipRestrictionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	restriction, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP restriction", err.Error())
		return
	}

	var model ipRestrictionDataSourceModel
	model.ID = types.StringValue(restriction.ID)
	model.URI = types.StringValue(restriction.URI)
	model.CreatedAt = types.StringValue(restriction.CreatedAt)
	model.Description = types.StringValue(restriction.Description)
	model.Metadata = types.StringValue(restriction.Metadata)
	model.Enforced = types.BoolValue(restriction.Enforced)
	model.Type = types.StringValue(restriction.Type)
	model.IPPolicyIDs = flattenRefList(restriction.IPPolicies)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
