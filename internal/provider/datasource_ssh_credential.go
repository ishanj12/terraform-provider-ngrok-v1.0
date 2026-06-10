package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/ssh_credentials"
)

var _ datasource.DataSource = &sshCredentialDataSource{}

type sshCredentialDataSourceModel struct {
	ID          types.String   `tfsdk:"id"`
	URI         types.String   `tfsdk:"uri"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	Description types.String   `tfsdk:"description"`
	Metadata    types.String   `tfsdk:"metadata"`
	PublicKey   types.String   `tfsdk:"public_key"`
	ACL         []types.String `tfsdk:"acl"`
	OwnerID     types.String   `tfsdk:"owner_id"`
}

type sshCredentialDataSource struct {
	client *ssh_credentials.Client
}

func NewSSHCredentialDataSource() datasource.DataSource {
	return &sshCredentialDataSource{}
}

func (d *sshCredentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_credential"
}

func (d *sshCredentialDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up an SSH Credential by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique SSH credential resource identifier.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the SSH credential API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the SSH credential was created, RFC 3339 format.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of who or what will use the SSH credential to authenticate.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this SSH credential.",
				Computed:    true,
			},
			"public_key": schema.StringAttribute{
				Description: "The PEM-encoded public key of the SSH keypair that will be used to authenticate.",
				Computed:    true,
			},
			"acl": schema.ListAttribute{
				Description: "Optional list of ACL rules.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"owner_id": schema.StringAttribute{
				Description: "The owner of this SSH credential.",
				Computed:    true,
			},
		},
	}
}

func (d *sshCredentialDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = ssh_credentials.NewClient(clientConfig)
}

func (d *sshCredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config sshCredentialDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cred, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading SSH credential", err.Error())
		return
	}

	var model sshCredentialDataSourceModel
	model.ID = types.StringValue(cred.ID)
	model.URI = types.StringValue(cred.URI)
	model.CreatedAt = types.StringValue(cred.CreatedAt)
	model.Description = types.StringValue(cred.Description)
	model.Metadata = types.StringValue(cred.Metadata)
	model.PublicKey = types.StringValue(cred.PublicKey)
	model.ACL = flattenStringList(cred.ACL)

	if cred.OwnerID != nil {
		model.OwnerID = types.StringValue(*cred.OwnerID)
	} else {
		model.OwnerID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
