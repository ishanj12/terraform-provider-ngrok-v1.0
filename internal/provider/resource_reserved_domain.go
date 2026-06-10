package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/reserved_domains"
)

var (
	_ resource.Resource                = &reservedDomainResource{}
	_ resource.ResourceWithImportState = &reservedDomainResource{}
)

type reservedDomainResourceModel struct {
	ID                          types.String   `tfsdk:"id"`
	Domain                      types.String   `tfsdk:"domain"`
	Region                      types.String   `tfsdk:"region"`
	Description                 types.String   `tfsdk:"description"`
	Metadata                    types.String   `tfsdk:"metadata"`
	CertificateID               types.String   `tfsdk:"certificate_id"`
	CertificateManagementPolicy types.Object   `tfsdk:"certificate_management_policy"`
	CNAMETarget                 types.String   `tfsdk:"cname_target"`
	ACMEChallengeCNAMETarget    types.String   `tfsdk:"acme_challenge_cname_target"`
	ResolvesTo                  []types.String `tfsdk:"resolves_to"`
	URI                         types.String   `tfsdk:"uri"`
	CreatedAt                   types.String   `tfsdk:"created_at"`
}

type reservedDomainResource struct {
	client *reserved_domains.Client
}

func NewReservedDomainResource() resource.Resource {
	return &reservedDomainResource{}
}

func (r *reservedDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reserved_domain"
}

func (r *reservedDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reserved Domains are hostnames that you can listen for traffic on. Domains can be used to listen for http, https or tls traffic. You may use a domain that you own by creating a CNAME record specified in the returned resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique reserved domain resource identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Hostname of the reserved domain.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description:       "Deprecated: With the launch of the ngrok Global Network domains traffic is now handled globally. This field applied only to endpoints.",
				DeprecationMessage: "This field is deprecated and will be removed in a future version.",
				Optional:          true,
				Computed:          true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of what this reserved domain will be used for.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this reserved domain. Optional, max 4096 bytes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_id": schema.StringAttribute{
				Description: "ID of a user-uploaded TLS certificate to use for connections to targeting this domain. Optional, mutually exclusive with certificate_management_policy.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_management_policy": schema.SingleNestedAttribute{
				Description: "Configuration for automatic management of TLS certificates for this domain, or null if automatic management is disabled. Mutually exclusive with certificate_id.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"authority": schema.StringAttribute{
						Description: "Certificate authority to request certificates from. The only supported value is letsencrypt.",
						Optional:    true,
						Computed:    true,
					},
					"private_key_type": schema.StringAttribute{
						Description: "Type of private key to use when requesting certificates. Defaults to ecdsa, can be either rsa or ecdsa.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"resolves_to": schema.ListAttribute{
				Description: "A list of ngrok point-of-presence shortcodes (or \"global\") that the domain resolves to.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"cname_target": schema.StringAttribute{
				Description: "DNS CNAME target for a custom hostname, or null if the reserved domain is a subdomain of an ngrok owned domain (e.g. *.ngrok.app).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"acme_challenge_cname_target": schema.StringAttribute{
				Description: "DNS CNAME target for the host _acme-challenge.example.com, where example.com is your reserved domain name. Required to issue certificates for wildcard, non-ngrok reserved domains.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				Description: "URI of the reserved domain API resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the reserved domain was created, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *reservedDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientConfig, ok := req.ProviderData.(*ngrok.ClientConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ngrok.ClientConfig, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = reserved_domains.NewClient(clientConfig)
}

func (r *reservedDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan reservedDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &ngrok.ReservedDomainCreate{
		Domain:      plan.Domain.ValueString(),
		Description: plan.Description.ValueString(),
		Metadata:    plan.Metadata.ValueString(),
	}
	if !plan.Region.IsNull() && !plan.Region.IsUnknown() {
		createReq.Region = plan.Region.ValueString()
	}
	createReq.CertificateID = stringPtrFromFramework(plan.CertificateID)
	createReq.CertificateManagementPolicy = expandCertPolicy(ctx, plan.CertificateManagementPolicy, &resp.Diagnostics)
	createReq.ResolvesTo = expandResolvesTo(plan.ResolvesTo)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating reserved domain", err.Error())
		return
	}

	flattenReservedDomain(ctx, domain, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reservedDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state reservedDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading reserved domain", err.Error())
		return
	}

	flattenReservedDomain(ctx, domain, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *reservedDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan reservedDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state reservedDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &ngrok.ReservedDomainUpdate{
		ID:          state.ID.ValueString(),
		Description: stringPtrFromFramework(plan.Description),
		Metadata:    stringPtrFromFramework(plan.Metadata),
	}
	updateReq.CertificateID = stringPtrFromFramework(plan.CertificateID)
	updateReq.CertificateManagementPolicy = expandCertPolicy(ctx, plan.CertificateManagementPolicy, &resp.Diagnostics)
	updateReq.ResolvesTo = expandResolvesTo(plan.ResolvesTo)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.Update(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating reserved domain", err.Error())
		return
	}

	flattenReservedDomain(ctx, domain, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reservedDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state reservedDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting reserved domain", err.Error())
	}
}

func (r *reservedDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func flattenReservedDomain(ctx context.Context, domain *ngrok.ReservedDomain, model *reservedDomainResourceModel, diags *diag.Diagnostics) {
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

	if len(domain.ResolvesTo) > 0 {
		model.ResolvesTo = flattenResolvesTo(domain.ResolvesTo)
	}

	// Only populate cert policy if user configured it or it was previously in state
	if !model.CertificateManagementPolicy.IsNull() && !model.CertificateManagementPolicy.IsUnknown() {
		model.CertificateManagementPolicy = flattenCertPolicy(ctx, domain.CertificateManagementPolicy, diags)
	} else if model.CertificateManagementPolicy.IsUnknown() {
		model.CertificateManagementPolicy = types.ObjectNull(certPolicyAttrTypes())
	}
}

func flattenCertPolicy(ctx context.Context, policy *ngrok.ReservedDomainCertPolicy, diags *diag.Diagnostics) types.Object {
	if policy == nil {
		return types.ObjectNull(certPolicyAttrTypes())
	}

	obj, d := types.ObjectValueFrom(ctx, certPolicyAttrTypes(), &certPolicyModel{
		Authority:      types.StringValue(policy.Authority),
		PrivateKeyType: types.StringValue(policy.PrivateKeyType),
	})
	diags.Append(d...)
	return obj
}

func expandCertPolicy(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *ngrok.ReservedDomainCertPolicy {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	var model certPolicyModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &ngrok.ReservedDomainCertPolicy{
		Authority:      model.Authority.ValueString(),
		PrivateKeyType: model.PrivateKeyType.ValueString(),
	}
}

type certPolicyModel struct {
	Authority      types.String `tfsdk:"authority"`
	PrivateKeyType types.String `tfsdk:"private_key_type"`
}

func certPolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"authority":        types.StringType,
		"private_key_type": types.StringType,
	}
}

func expandResolvesTo(vals []types.String) []ngrok.ReservedDomainResolvesToEntry {
	if vals == nil {
		return nil
	}
	entries := make([]ngrok.ReservedDomainResolvesToEntry, len(vals))
	for i, v := range vals {
		entries[i] = ngrok.ReservedDomainResolvesToEntry{Value: v.ValueString()}
	}
	return entries
}

func flattenResolvesTo(entries []ngrok.ReservedDomainResolvesToEntry) []types.String {
	if entries == nil {
		return nil
	}
	result := make([]types.String, len(entries))
	for i, e := range entries {
		result[i] = types.StringValue(e.Value)
	}
	return result
}
