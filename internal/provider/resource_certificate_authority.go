package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/certificate_authorities"
)

var (
	_ resource.Resource                = &certificateAuthorityResource{}
	_ resource.ResourceWithImportState = &certificateAuthorityResource{}
)

type certificateAuthorityResourceModel struct {
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

type certificateAuthorityResource struct {
	client *certificate_authorities.Client
}

func NewCertificateAuthorityResource() resource.Resource {
	return &certificateAuthorityResource{}
}

func (r *certificateAuthorityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_authority"
}

func (r *certificateAuthorityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Certificate Authorities are x509 certificates that are used to sign other x509 certificates. Attach a Certificate Authority to the Mutual TLS module to verify that the TLS certificate presented by a client has been signed by this CA.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this Certificate Authority.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				Description: "URI of the Certificate Authority API resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the Certificate Authority was created, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this Certificate Authority. Optional, max 255 bytes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this Certificate Authority. Optional, max 4096 bytes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ca_pem": schema.StringAttribute{
				Description: "Raw PEM of the Certificate Authority.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subject_common_name": schema.StringAttribute{
				Description: "Subject common name of the Certificate Authority.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"not_before": schema.StringAttribute{
				Description: "Timestamp when this Certificate Authority becomes valid, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"not_after": schema.StringAttribute{
				Description: "Timestamp when this Certificate Authority becomes invalid, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_usages": schema.ListAttribute{
				Description: "Set of actions the private key of this Certificate Authority can be used for.",
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"extended_key_usages": schema.ListAttribute{
				Description: "Extended set of actions the private key of this Certificate Authority can be used for.",
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *certificateAuthorityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = certificate_authorities.NewClient(clientConfig)
}

func (r *certificateAuthorityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan certificateAuthorityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &ngrok.CertificateAuthorityCreate{
		Description: plan.Description.ValueString(),
		Metadata:    plan.Metadata.ValueString(),
		CAPEM:       plan.CAPEM.ValueString(),
	}

	ca, err := r.client.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating certificate authority", err.Error())
		return
	}

	flattenCertificateAuthority(ca, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *certificateAuthorityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state certificateAuthorityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ca, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading certificate authority", err.Error())
		return
	}

	flattenCertificateAuthority(ca, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *certificateAuthorityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan certificateAuthorityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state certificateAuthorityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &ngrok.CertificateAuthorityUpdate{
		ID:          state.ID.ValueString(),
		Description: stringPtrFromFramework(plan.Description),
		Metadata:    stringPtrFromFramework(plan.Metadata),
	}

	ca, err := r.client.Update(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating certificate authority", err.Error())
		return
	}

	flattenCertificateAuthority(ca, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *certificateAuthorityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state certificateAuthorityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting certificate authority", err.Error())
	}
}

func (r *certificateAuthorityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func flattenCertificateAuthority(ca *ngrok.CertificateAuthority, model *certificateAuthorityResourceModel) {
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
