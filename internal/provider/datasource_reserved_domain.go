package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/reserved_domains"
)

var _ datasource.DataSource = &reservedDomainDataSource{}

type reservedDomainDataSourceModel struct {
	ID                          types.String `tfsdk:"id"`
	Domain                      types.String `tfsdk:"domain"`
	Region                      types.String `tfsdk:"region"`
	Description                 types.String `tfsdk:"description"`
	Metadata                    types.String `tfsdk:"metadata"`
	CertificateID               types.String `tfsdk:"certificate_id"`
	CertificateManagementPolicy types.Object `tfsdk:"certificate_management_policy"`
	CNAMETarget                 types.String `tfsdk:"cname_target"`
	ACMEChallengeCNAMETarget    types.String `tfsdk:"acme_challenge_cname_target"`
	URI                         types.String `tfsdk:"uri"`
	CreatedAt                   types.String `tfsdk:"created_at"`
}

type reservedDomainDataSource struct {
	client *reserved_domains.Client
}

func NewReservedDomainDataSource() datasource.DataSource {
	return &reservedDomainDataSource{}
}

func (d *reservedDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reserved_domain"
}

func (d *reservedDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up a reserved domain by ID or domain name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique reserved domain resource identifier. Provide either id or domain.",
				Optional:    true,
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "Hostname of the reserved domain. Provide either id or domain.",
				Optional:    true,
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Deprecated region field.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of what this reserved domain will be used for.",
				Computed:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this reserved domain.",
				Computed:    true,
			},
			"certificate_id": schema.StringAttribute{
				Description: "ID of a user-uploaded TLS certificate used for connections to this domain.",
				Computed:    true,
			},
			"certificate_management_policy": schema.SingleNestedAttribute{
				Description: "Configuration for automatic management of TLS certificates for this domain.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"authority": schema.StringAttribute{
						Description: "Certificate authority to request certificates from.",
						Computed:    true,
					},
					"private_key_type": schema.StringAttribute{
						Description: "Type of private key to use when requesting certificates.",
						Computed:    true,
					},
				},
			},
			"cname_target": schema.StringAttribute{
				Description: "DNS CNAME target for a custom hostname.",
				Computed:    true,
			},
			"acme_challenge_cname_target": schema.StringAttribute{
				Description: "DNS CNAME target for the host _acme-challenge.example.com.",
				Computed:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the reserved domain API resource.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the reserved domain was created, RFC 3339 format.",
				Computed:    true,
			},
		},
	}
}

func (d *reservedDomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = reserved_domains.NewClient(clientConfig)
}

func (d *reservedDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config reservedDomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var domain *ngrok.ReservedDomain

	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		// Lookup by ID
		var err error
		domain, err = d.client.Get(ctx, config.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading reserved domain", err.Error())
			return
		}
	} else if !config.Domain.IsNull() && config.Domain.ValueString() != "" {
		// Lookup by domain name using CEL filter
		filter := fmt.Sprintf(`obj.domain == %q`, config.Domain.ValueString())
		iter := d.client.List(&ngrok.FilteredPaging{
			Filter: &filter,
		})

		if !iter.Next(ctx) {
			if err := iter.Err(); err != nil {
				resp.Diagnostics.AddError("Error listing reserved domains", err.Error())
				return
			}
			resp.Diagnostics.AddError(
				"Reserved domain not found",
				fmt.Sprintf("No reserved domain found with domain %q.", config.Domain.ValueString()),
			)
			return
		}
		domain = iter.Item()

		// Ensure there is only one match
		if iter.Next(ctx) {
			resp.Diagnostics.AddError(
				"Multiple reserved domains found",
				fmt.Sprintf("More than one reserved domain found with domain %q. Use id instead.", config.Domain.ValueString()),
			)
			return
		}
		if err := iter.Err(); err != nil {
			resp.Diagnostics.AddError("Error listing reserved domains", err.Error())
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Missing lookup attribute",
			"Either id or domain must be specified.",
		)
		return
	}

	var model reservedDomainDataSourceModel
	flattenReservedDomainDataSource(ctx, domain, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenReservedDomainDataSource(ctx context.Context, domain *ngrok.ReservedDomain, model *reservedDomainDataSourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(domain.ID)
	model.Domain = types.StringValue(domain.Domain)
	model.Region = types.StringValue(domain.Region)
	model.Description = types.StringValue(domain.Description)
	model.Metadata = types.StringValue(domain.Metadata)
	model.URI = types.StringValue(domain.URI)
	model.CreatedAt = types.StringValue(domain.CreatedAt)

	if domain.CNAMETarget != nil {
		model.CNAMETarget = types.StringValue(*domain.CNAMETarget)
	} else {
		model.CNAMETarget = types.StringNull()
	}

	if domain.ACMEChallengeCNAMETarget != nil {
		model.ACMEChallengeCNAMETarget = types.StringValue(*domain.ACMEChallengeCNAMETarget)
	} else {
		model.ACMEChallengeCNAMETarget = types.StringNull()
	}

	model.CertificateID = types.StringValue(flattenRef(domain.Certificate))

	model.CertificateManagementPolicy = flattenCertPolicy(ctx, domain.CertificateManagementPolicy, diags)
}
