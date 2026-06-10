package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/ip_restrictions"
)

var (
	_ resource.Resource                = &ipRestrictionResource{}
	_ resource.ResourceWithImportState = &ipRestrictionResource{}
)

type ipRestrictionResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	URI         types.String   `tfsdk:"uri"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	Description types.String   `tfsdk:"description"`
	Metadata    types.String   `tfsdk:"metadata"`
	Enforced    types.Bool     `tfsdk:"enforced"`
	Type        types.String   `tfsdk:"type"`
	IPPolicyIDs []types.String `tfsdk:"ip_policy_ids"`
}

type ipRestrictionResource struct {
	client *ip_restrictions.Client
}

func NewIPRestrictionResource() resource.Resource {
	return &ipRestrictionResource{}
}

func (r *ipRestrictionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_restriction"
}

func (r *ipRestrictionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An IP restriction is a restriction placed on the CIDRs that are allowed to initiate traffic to a specific aspect of your ngrok accounts. It uses IP policies to define which CIDRs are allowed and denied.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this IP restriction.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				Description: "URI of the IP restriction API resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the IP restriction was created, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this IP restriction.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this IP restriction. Optional, max 4096 bytes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enforced": schema.BoolAttribute{
				Description: "True if the IP restriction will be enforced. If false, only combinator will be active effect.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"type": schema.StringAttribute{
				Description: "Type of the IP restriction. This defines what traffic will be restricted with the attached policies. Accepted values are dashboard, api, agent, or endpoints.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_policy_ids": schema.ListAttribute{
				Description: "The set of IP policy identifiers that are used to enforce the restriction.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *ipRestrictionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = ip_restrictions.NewClient(clientConfig)
}

func (r *ipRestrictionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	restriction, err := r.client.Create(ctx, &ngrok.IPRestrictionCreate{
		Description: plan.Description.ValueString(),
		Metadata:    plan.Metadata.ValueString(),
		Enforced:    plan.Enforced.ValueBool(),
		Type:        plan.Type.ValueString(),
		IPPolicyIDs: expandStringList(plan.IPPolicyIDs),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating IP restriction", err.Error())
		return
	}

	flattenIPRestriction(restriction, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipRestrictionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	restriction, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading IP restriction", err.Error())
		return
	}

	flattenIPRestriction(restriction, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipRestrictionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ipRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ipRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	restriction, err := r.client.Update(ctx, &ngrok.IPRestrictionUpdate{
		ID:          state.ID.ValueString(),
		Description: stringPtrFromFramework(plan.Description),
		Metadata:    stringPtrFromFramework(plan.Metadata),
		Enforced:    boolPtrFromFramework(plan.Enforced),
		IPPolicyIDs: expandStringList(plan.IPPolicyIDs),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating IP restriction", err.Error())
		return
	}

	flattenIPRestriction(restriction, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipRestrictionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting IP restriction", err.Error())
	}
}

func (r *ipRestrictionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func flattenIPRestriction(restriction *ngrok.IPRestriction, model *ipRestrictionResourceModel) {
	model.ID = types.StringValue(restriction.ID)
	model.URI = types.StringValue(restriction.URI)
	model.CreatedAt = types.StringValue(restriction.CreatedAt)
	model.Description = types.StringValue(restriction.Description)
	model.Metadata = types.StringValue(restriction.Metadata)
	model.Enforced = types.BoolValue(restriction.Enforced)
	model.Type = types.StringValue(restriction.Type)
	// Preserve the user's configured order for ip_policy_ids.
	// The API may return refs in a different order; reorder to match
	// the prior state/plan to avoid spurious diffs.
	apiIDs := flattenRefList(restriction.IPPolicies)
	if model.IPPolicyIDs != nil {
		model.IPPolicyIDs = reorderStringList(model.IPPolicyIDs, apiIDs)
	} else {
		model.IPPolicyIDs = apiIDs
	}
}

// reorderStringList returns apiVals reordered to match the order in prior,
// with any new values appended at the end and removed values dropped.
func reorderStringList(prior, apiVals []types.String) []types.String {
	apiSet := make(map[string]bool, len(apiVals))
	for _, v := range apiVals {
		apiSet[v.ValueString()] = true
	}
	var result []types.String
	// First, add items from prior that still exist in API response
	for _, v := range prior {
		if apiSet[v.ValueString()] {
			result = append(result, v)
			delete(apiSet, v.ValueString())
		}
	}
	// Then append any new items from API not in prior
	for _, v := range apiVals {
		if apiSet[v.ValueString()] {
			result = append(result, v)
		}
	}
	return result
}
