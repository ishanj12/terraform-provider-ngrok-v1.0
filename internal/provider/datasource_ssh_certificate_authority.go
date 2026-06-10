package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/ssh_certificate_authorities"
)

var _ datasource.DataSource = &sshCertificateAuthorityDataSource{}

type sshCertificateAuthorityDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	URI         types.String `tfsdk:"uri"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	Metadata    types.String `tfsdk:"metadata"`
	PublicKey   types.String `tfsdk:"public_key"`
	KeyType     types.String `tfsdk:"key_type"`
}

type sshCertificateAuthorityDataSource struct {
	client *ssh_certificate_authorities.Client
}

func NewSSHCertificateAuthorityDataSource() datasource.DataSource {
	return &sshCertificateAuthorityDataSource{}
}

func (d *sshCertificateAuthorityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_certificate_authority"
}

func (d *sshCertificateAuthorityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up an SSH Certificate Authority by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this SSH Certificate Authority.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the SSH Certificate Authority API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the SSH Certificate Authority was created, RFC 3339 format.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this SSH Certificate Authority.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this SSH Certificate Authority.",
				Computed:    true,
			},
			"public_key": schema.StringAttribute{
				Description: "Raw public key of this SSH Certificate Authority.",
				Computed:    true,
			},
			"key_type": schema.StringAttribute{
				Description: "The type of private key for this SSH Certificate Authority.",
				Computed:    true,
			},
		},
	}
}

func (d *sshCertificateAuthorityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = ssh_certificate_authorities.NewClient(clientConfig)
}

func (d *sshCertificateAuthorityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config sshCertificateAuthorityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ca, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading SSH certificate authority", err.Error())
		return
	}

	var model sshCertificateAuthorityDataSourceModel
	model.ID = types.StringValue(ca.ID)
	model.URI = types.StringValue(ca.URI)
	model.CreatedAt = types.StringValue(ca.CreatedAt)
	model.Description = types.StringValue(ca.Description)
	model.Metadata = types.StringValue(ca.Metadata)
	model.PublicKey = types.StringValue(ca.PublicKey)
	model.KeyType = types.StringValue(ca.KeyType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
