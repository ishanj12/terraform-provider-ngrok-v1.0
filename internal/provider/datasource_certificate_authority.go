package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/certificate_authorities"
)

var _ datasource.DataSource = &certificateAuthorityDataSource{}

type certificateAuthorityDataSourceModel struct {
	ID                types.String   `tfsdk:"id"`
	URI               types.String   `tfsdk:"uri"`
	CreatedAt         types.String   `tfsdk:"created_at"`
	Description       types.String   `tfsdk:"description"`
	Metadata          types.String   `tfsdk:"metadata"`
	CAPEM             types.String   `tfsdk:"ca_pem"`
	SubjectCommonName types.String   `tfsdk:"subject_common_name"`
	NotBefore         types.String   `tfsdk:"not_before"`
	NotAfter          types.String   `tfsdk:"not_after"`
	KeyUsages         []types.String `tfsdk:"key_usages"`
	ExtendedKeyUsages []types.String `tfsdk:"extended_key_usages"`
}

type certificateAuthorityDataSource struct {
	client *certificate_authorities.Client
}

func NewCertificateAuthorityDataSource() datasource.DataSource {
	return &certificateAuthorityDataSource{}
}

func (d *certificateAuthorityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_authority"
}

func (d *certificateAuthorityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a Certificate Authority by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this Certificate Authority.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the Certificate Authority API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the Certificate Authority was created, RFC 3339 format.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this Certificate Authority.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this Certificate Authority.",
				Computed:    true,
			},
			"ca_pem": schema.StringAttribute{
				Description: "Raw PEM of the Certificate Authority.",
				Computed:    true,
			},
			"subject_common_name": schema.StringAttribute{
				Description: "Subject common name of the Certificate Authority.",
				Computed:    true,
			},
			"not_before": schema.StringAttribute{
				Description: "Timestamp when this Certificate Authority becomes valid, RFC 3339 format.",
				Computed:    true,
			},
			"not_after": schema.StringAttribute{
				Description: "Timestamp when this Certificate Authority becomes invalid, RFC 3339 format.",
				Computed:    true,
			},
			"key_usages": schema.ListAttribute{
				Description: "Set of actions the private key of this Certificate Authority can be used for.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"extended_key_usages": schema.ListAttribute{
				Description: "Extended set of actions the private key of this Certificate Authority can be used for.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *certificateAuthorityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = certificate_authorities.NewClient(clientConfig)
}

func (d *certificateAuthorityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config certificateAuthorityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ca, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading certificate authority", err.Error())
		return
	}

	var model certificateAuthorityDataSourceModel
	flattenCertificateAuthorityDataSource(ca, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenCertificateAuthorityDataSource(ca *ngrok.CertificateAuthority, model *certificateAuthorityDataSourceModel) {
	model.ID = types.StringValue(ca.ID)
	model.URI = types.StringValue(ca.URI)
	model.CreatedAt = types.StringValue(ca.CreatedAt)
	model.Description = types.StringValue(ca.Description)
	model.Metadata = types.StringValue(ca.Metadata)
	model.CAPEM = types.StringValue(ca.CAPEM)
	model.SubjectCommonName = types.StringValue(ca.SubjectCommonName)
	model.NotBefore = types.StringValue(ca.NotBefore)
	model.NotAfter = types.StringValue(ca.NotAfter)
	model.KeyUsages = flattenStringList(ca.KeyUsages)
	model.ExtendedKeyUsages = flattenStringList(ca.ExtendedKeyUsages)
}
