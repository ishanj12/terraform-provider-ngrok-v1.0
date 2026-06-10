package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/endpoints"
)

var (
	_ resource.Resource                = &cloudEndpointResource{}
	_ resource.ResourceWithImportState = &cloudEndpointResource{}
	_ resource.ResourceWithModifyPlan  = &cloudEndpointResource{}
)

type cloudEndpointResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	URL            types.String   `tfsdk:"url"`
	Type           types.String   `tfsdk:"type"`
	TrafficPolicy  types.String   `tfsdk:"traffic_policy"`
	Description    types.String   `tfsdk:"description"`
	Metadata       types.String   `tfsdk:"metadata"`
	Bindings       []types.String `tfsdk:"bindings"`
	PoolingEnabled types.Bool     `tfsdk:"pooling_enabled"`
	DomainID       types.String   `tfsdk:"domain_id"`
	Region         types.String   `tfsdk:"region"`
	URI            types.String   `tfsdk:"uri"`
	CreatedAt      types.String   `tfsdk:"created_at"`
	UpdatedAt      types.String   `tfsdk:"updated_at"`
}

type cloudEndpointResource struct {
	client *endpoints.Client
}

func NewCloudEndpointResource() resource.Resource {
	return &cloudEndpointResource{}
}

func (r *cloudEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_endpoint"
}

func (r *cloudEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Cloud Endpoints are endpoints that are created and managed by the ngrok cloud. They can be used to route traffic to your services using traffic policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique endpoint resource identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Description: "The URL of the endpoint.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of endpoint. Always \"cloud\" for cloud endpoints.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"traffic_policy": schema.StringAttribute{
				Description: "The traffic policy attached to this endpoint. Must be valid JSON.",
				Required:    true,
				Validators: []validator.String{
					JSONSyntax(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this cloud endpoint.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this cloud endpoint. Optional, max 4096 bytes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bindings": schema.ListAttribute{
				Description: "The bindings associated with this endpoint. Defaults to [\"public\"].",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default: listdefault.StaticValue(
					types.ListValueMust(types.StringType, []attr.Value{types.StringValue("public")}),
				),
			},
			"pooling_enabled": schema.BoolAttribute{
				Description: "Whether the endpoint allows connection pooling.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"domain_id": schema.StringAttribute{
				Description: "ID of the domain reserved for this endpoint.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Description: "Region of the endpoint.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				Description: "URI of the cloud endpoint API resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the endpoint was created, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the endpoint was last updated, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *cloudEndpointResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on create or destroy
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan cloudEndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state cloudEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If any user-configurable field changed, mark updated_at as unknown
	// so the provider can return the new value from the API.
	hasChanges := !plan.Description.Equal(state.Description) ||
		!plan.Metadata.Equal(state.Metadata) ||
		!plan.URL.Equal(state.URL) ||
		!plan.TrafficPolicy.Equal(state.TrafficPolicy) ||
		!plan.PoolingEnabled.Equal(state.PoolingEnabled) ||
		!stringListEqual(plan.Bindings, state.Bindings)

	if hasChanges {
		resp.Plan.SetAttribute(ctx, path.Root("updated_at"), types.StringUnknown())
	}
}

func (r *cloudEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = endpoints.NewClient(clientConfig)
}

func (r *cloudEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudEndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &ngrok.EndpointCreate{
		URL:            plan.URL.ValueString(),
		TrafficPolicy:  plan.TrafficPolicy.ValueString(),
		Description:    stringPtrFromFramework(plan.Description),
		Metadata:       stringPtrFromFramework(plan.Metadata),
		Bindings:       expandStringList(plan.Bindings),
		PoolingEnabled: boolPtrFromFramework(plan.PoolingEnabled),
	}

	endpoint, err := r.client.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cloud endpoint", err.Error())
		return
	}

	flattenCloudEndpoint(endpoint, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading cloud endpoint", err.Error())
		return
	}

	flattenCloudEndpoint(endpoint, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cloudEndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state cloudEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &ngrok.EndpointUpdate{
		ID:             state.ID.ValueString(),
		TrafficPolicy:  stringPtrFromFramework(plan.TrafficPolicy),
		Description:    stringPtrFromFramework(plan.Description),
		Metadata:       stringPtrFromFramework(plan.Metadata),
		Bindings:       expandStringList(plan.Bindings),
		PoolingEnabled: boolPtrFromFramework(plan.PoolingEnabled),
	}

	endpoint, err := r.client.Update(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating cloud endpoint", err.Error())
		return
	}

	flattenCloudEndpoint(endpoint, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudEndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting cloud endpoint", err.Error())
	}
}

func (r *cloudEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func flattenCloudEndpoint(endpoint *ngrok.Endpoint, model *cloudEndpointResourceModel) {
	model.ID = types.StringValue(endpoint.ID)
	model.URL = types.StringValue(endpoint.URL)
	model.Type = types.StringValue(endpoint.Type)
	model.TrafficPolicy = types.StringValue(endpoint.TrafficPolicy)
	model.Description = types.StringValue(endpoint.Description)
	model.Metadata = types.StringValue(endpoint.Metadata)
	model.Bindings = flattenStringList(endpoint.Bindings)
	model.PoolingEnabled = types.BoolValue(endpoint.PoolingEnabled)
	model.DomainID = types.StringValue(flattenRef(endpoint.Domain))
	model.Region = types.StringValue(endpoint.Region)
	model.URI = types.StringValue(endpoint.URI)
	model.CreatedAt = types.StringValue(endpoint.CreatedAt)
	model.UpdatedAt = types.StringValue(endpoint.UpdatedAt)
}
