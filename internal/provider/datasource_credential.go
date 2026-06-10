package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/credentials"
)

var _ datasource.DataSource = &credentialDataSource{}

type credentialDataSourceModel struct {
	ID          types.String   `tfsdk:"id"`
	URI         types.String   `tfsdk:"uri"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	Description types.String   `tfsdk:"description"`
	Metadata    types.String   `tfsdk:"metadata"`
	ACL         []types.String `tfsdk:"acl"`
	OwnerID     types.String   `tfsdk:"owner_id"`
}

type credentialDataSource struct {
	client *credentials.Client
}

func NewCredentialDataSource() datasource.DataSource {
	return &credentialDataSource{}
}

func (d *credentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (d *credentialDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a tunnel credential by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique tunnel credential resource identifier.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the tunnel credential API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the tunnel credential was created, RFC 3339 format.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of what this credential is used for.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this credential.",
				Computed:    true,
			},
			"acl": schema.ListAttribute{
				Description: "List of ACL rules applied to this credential.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"owner_id": schema.StringAttribute{
				Description: "The owner ID of the credential.",
				Computed:    true,
			},
		},
	}
}

func (d *credentialDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = credentials.NewClient(clientConfig)
}

func (d *credentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config credentialDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cred, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading credential", err.Error())
		return
	}

	var model credentialDataSourceModel
	model.ID = types.StringValue(cred.ID)
	model.URI = types.StringValue(cred.URI)
	model.CreatedAt = types.StringValue(cred.CreatedAt)
	model.Description = types.StringValue(cred.Description)
	model.Metadata = types.StringValue(cred.Metadata)
	model.ACL = flattenStringList(cred.ACL)

	if cred.OwnerID != nil {
		model.OwnerID = types.StringValue(*cred.OwnerID)
	} else {
		model.OwnerID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
