package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/tls_certificates"
)

var _ datasource.DataSource = &tlsCertificateDataSource{}

type tlsCertificateDataSourceModel struct {
	ID                        types.String   `tfsdk:"id"`
	URI                       types.String   `tfsdk:"uri"`
	CreatedAt                 types.String   `tfsdk:"created_at"`
	Description               types.String   `tfsdk:"description"`
	Metadata                  types.String   `tfsdk:"metadata"`
	CertificatePEM            types.String   `tfsdk:"certificate_pem"`
	SubjectCommonName         types.String   `tfsdk:"subject_common_name"`
	SubjectAlternativeNames   types.Object   `tfsdk:"subject_alternative_names"`
	IssuedAt                  types.String   `tfsdk:"issued_at"`
	NotBefore                 types.String   `tfsdk:"not_before"`
	NotAfter                  types.String   `tfsdk:"not_after"`
	KeyUsages                 []types.String `tfsdk:"key_usages"`
	ExtendedKeyUsages         []types.String `tfsdk:"extended_key_usages"`
	PrivateKeyType            types.String   `tfsdk:"private_key_type"`
	IssuerCommonName          types.String   `tfsdk:"issuer_common_name"`
	SerialNumber              types.String   `tfsdk:"serial_number"`
	SubjectOrganization       types.String   `tfsdk:"subject_organization"`
	SubjectOrganizationalUnit types.String   `tfsdk:"subject_organizational_unit"`
	SubjectLocality           types.String   `tfsdk:"subject_locality"`
	SubjectProvince           types.String   `tfsdk:"subject_province"`
	SubjectCountry            types.String   `tfsdk:"subject_country"`
}

type tlsCertificateDataSource struct {
	client *tls_certificates.Client
}

func NewTLSCertificateDataSource() datasource.DataSource {
	return &tlsCertificateDataSource{}
}

func (d *tlsCertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tls_certificate"
}

func (d *tlsCertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a TLS certificate by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this TLS certificate.",
				Required:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the TLS certificate API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the TLS certificate was created, RFC 3339 format.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this TLS certificate.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this TLS certificate.",
				Computed:    true,
			},
			"certificate_pem": schema.StringAttribute{
				Description: "Chain of PEM-encoded certificates, leaf first.",
				Computed:    true,
			},
			"subject_common_name": schema.StringAttribute{
				Description: "Subject common name from the leaf of this TLS certificate.",
				Computed:    true,
			},
			"subject_alternative_names": schema.SingleNestedAttribute{
				Description: "Subject alternative names (SANs) from the leaf of this TLS certificate.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"dns_names": schema.ListAttribute{
						Description: "Set of additional domains (including wildcards) this TLS certificate is valid for.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"ips": schema.ListAttribute{
						Description: "Set of IP addresses this TLS certificate is also valid for.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"issued_at": schema.StringAttribute{
				Description: "Timestamp when this TLS certificate was issued automatically, or empty if user-uploaded.",
				Computed:    true,
			},
			"not_before": schema.StringAttribute{
				Description: "Timestamp when this TLS certificate becomes valid, RFC 3339 format.",
				Computed:    true,
			},
			"not_after": schema.StringAttribute{
				Description: "Timestamp when this TLS certificate becomes invalid, RFC 3339 format.",
				Computed:    true,
			},
			"key_usages": schema.ListAttribute{
				Description: "Set of actions the private key of this TLS certificate can be used for.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"extended_key_usages": schema.ListAttribute{
				Description: "Extended set of actions the private key of this TLS certificate can be used for.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"private_key_type": schema.StringAttribute{
				Description: "Type of the private key of this TLS certificate. One of rsa, ecdsa, or ed25519.",
				Computed:    true,
			},
			"issuer_common_name": schema.StringAttribute{
				Description: "Issuer common name from the leaf of this TLS certificate.",
				Computed:    true,
			},
			"serial_number": schema.StringAttribute{
				Description: "Serial number of the leaf of this TLS certificate.",
				Computed:    true,
			},
			"subject_organization": schema.StringAttribute{
				Description: "Subject organization from the leaf of this TLS certificate.",
				Computed:    true,
			},
			"subject_organizational_unit": schema.StringAttribute{
				Description: "Subject organizational unit from the leaf of this TLS certificate.",
				Computed:    true,
			},
			"subject_locality": schema.StringAttribute{
				Description: "Subject locality from the leaf of this TLS certificate.",
				Computed:    true,
			},
			"subject_province": schema.StringAttribute{
				Description: "Subject province from the leaf of this TLS certificate.",
				Computed:    true,
			},
			"subject_country": schema.StringAttribute{
				Description: "Subject country from the leaf of this TLS certificate.",
				Computed:    true,
			},
		},
	}
}

func (d *tlsCertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = tls_certificates.NewClient(clientConfig)
}

func (d *tlsCertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tlsCertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := d.client.Get(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading TLS certificate", err.Error())
		return
	}

	var model tlsCertificateDataSourceModel
	flattenTLSCertificateDataSource(ctx, cert, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenTLSCertificateDataSource(ctx context.Context, cert *ngrok.TLSCertificate, model *tlsCertificateDataSourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(cert.ID)
	model.URI = types.StringValue(cert.URI)
	model.CreatedAt = types.StringValue(cert.CreatedAt)
	model.Description = types.StringValue(cert.Description)
	model.Metadata = types.StringValue(cert.Metadata)
	model.CertificatePEM = types.StringValue(cert.CertificatePEM)
	model.SubjectCommonName = types.StringValue(cert.SubjectCommonName)
	model.SubjectAlternativeNames = flattenSANs(ctx, cert.SubjectAlternativeNames, diags)

	if cert.IssuedAt != nil {
		model.IssuedAt = types.StringValue(*cert.IssuedAt)
	} else {
		model.IssuedAt = types.StringNull()
	}

	model.NotBefore = types.StringValue(cert.NotBefore)
	model.NotAfter = types.StringValue(cert.NotAfter)
	model.KeyUsages = flattenStringList(cert.KeyUsages)
	model.ExtendedKeyUsages = flattenStringList(cert.ExtendedKeyUsages)
	model.PrivateKeyType = types.StringValue(cert.PrivateKeyType)
	model.IssuerCommonName = types.StringValue(cert.IssuerCommonName)
	model.SerialNumber = types.StringValue(cert.SerialNumber)
	model.SubjectOrganization = types.StringValue(cert.SubjectOrganization)
	model.SubjectOrganizationalUnit = types.StringValue(cert.SubjectOrganizationalUnit)
	model.SubjectLocality = types.StringValue(cert.SubjectLocality)
	model.SubjectProvince = types.StringValue(cert.SubjectProvince)
	model.SubjectCountry = types.StringValue(cert.SubjectCountry)
}
