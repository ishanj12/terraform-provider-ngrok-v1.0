package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/ip_policy_rules"
)

var _ datasource.DataSource = &ipPolicyRuleDataSource{}

type ipPolicyRuleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	URI         types.String `tfsdk:"uri"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	Metadata    types.String `tfsdk:"metadata"`
	CIDR        types.String `tfsdk:"cidr"`
	IPPolicyID  types.String `tfsdk:"ip_policy_id"`
	Action      types.String `tfsdk:"action"`
}

type ipPolicyRuleDataSource struct {
	client *ip_policy_rules.Client
}

func NewIPPolicyRuleDataSource() datasource.DataSource {
	return &ipPolicyRuleDataSource{}
}

func (d *ipPolicyRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_policy_rule"
}

func (d *ipPolicyRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up an IP policy rule by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this IP policy rule.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the IP policy rule API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the IP policy rule was created, RFC 3339 format.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of what this IP policy rule will be used for.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this IP policy rule.",
				Computed:    true,
			},
			"cidr": schema.StringAttribute{
				Description: "An IP or IP range specified in CIDR notation.",
				Computed:    true,
			},
			"ip_policy_id": schema.StringAttribute{
				Description: "ID of the IP policy this rule belongs to.",
				Computed:    true,
			},
			"action": schema.StringAttribute{
				Description: "The action to apply to the policy rule, either allow or deny.",
				Computed:    true,
			},
		},
	}
}

func (d *ipPolicyRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = ip_policy_rules.NewClient(clientConfig)
}

func (d *ipPolicyRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ipPolicyRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP policy rule", err.Error())
		return
	}

	var model ipPolicyRuleDataSourceModel
	model.ID = types.StringValue(rule.ID)
	model.URI = types.StringValue(rule.URI)
	model.CreatedAt = types.StringValue(rule.CreatedAt)
	model.Description = types.StringValue(rule.Description)
	model.Metadata = types.StringValue(rule.Metadata)
	model.CIDR = types.StringValue(rule.CIDR)
	model.IPPolicyID = types.StringValue(rule.IPPolicy.ID)
	model.Action = types.StringValue(rule.Action)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
