package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ngrok "github.com/ngrok/ngrok-api-go/v9"
	"github.com/ngrok/ngrok-api-go/v9/event_subscriptions"
)

var (
	_ resource.Resource                = &eventSubscriptionResource{}
	_ resource.ResourceWithImportState = &eventSubscriptionResource{}
)

type eventSubscriptionResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Description    types.String `tfsdk:"description"`
	Metadata       types.String `tfsdk:"metadata"`
	Sources        types.List   `tfsdk:"sources"`
	DestinationIDs types.List   `tfsdk:"destination_ids"`
	URI            types.String `tfsdk:"uri"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

type eventSubscriptionResource struct {
	client *event_subscriptions.Client
}

func NewEventSubscriptionResource() resource.Resource {
	return &eventSubscriptionResource{}
}

func (r *eventSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_subscription"
}

func (r *eventSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Event Subscriptions allow you to subscribe to events and send them to Event Destinations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique event subscription resource identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Arbitrary customer supplied information intended to be human readable. Optional, max 255 chars.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary customer supplied information intended to be machine readable. Optional, max 4096 chars.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sources": schema.ListNestedAttribute{
				Description: "Sources containing the types for which this event subscription will trigger.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Type of event for which an event subscription will trigger.",
							Required:    true,
						},
						"uri": schema.StringAttribute{
							Description: "URI of the Event Source API resource.",
							Computed:    true,
						},
					},
				},
			},
			"destination_ids": schema.ListAttribute{
				Description: "A list of Event Destination IDs which should be used for this Event Subscription.",
				Required:    true,
				ElementType: types.StringType,
			},
			"uri": schema.StringAttribute{
				Description: "URI of the Event Subscription API resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the Event Subscription was created, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *eventSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = event_subscriptions.NewClient(clientConfig)
}

func (r *eventSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan eventSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sources := expandEventSources(ctx, plan.Sources, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var destIDs []types.String
	resp.Diagnostics.Append(plan.DestinationIDs.ElementsAs(ctx, &destIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &ngrok.EventSubscriptionCreate{
		Description:    plan.Description.ValueString(),
		Metadata:       plan.Metadata.ValueString(),
		Sources:        sources,
		DestinationIDs: expandStringList(destIDs),
	}

	sub, err := r.client.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating event subscription", err.Error())
		return
	}

	flattenEventSubscription(ctx, sub, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state eventSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading event subscription", err.Error())
		return
	}

	flattenEventSubscription(ctx, sub, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *eventSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan eventSubscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state eventSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sources := expandEventSources(ctx, plan.Sources, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var destIDs []types.String
	resp.Diagnostics.Append(plan.DestinationIDs.ElementsAs(ctx, &destIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &ngrok.EventSubscriptionUpdate{
		ID:             state.ID.ValueString(),
		Description:    stringPtrFromFramework(plan.Description),
		Metadata:       stringPtrFromFramework(plan.Metadata),
		Sources:        sources,
		DestinationIDs: expandStringList(destIDs),
	}

	sub, err := r.client.Update(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating event subscription", err.Error())
		return
	}

	flattenEventSubscription(ctx, sub, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state eventSubscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting event subscription", err.Error())
	}
}

func (r *eventSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type eventSourceModel struct {
	Type types.String `tfsdk:"type"`
	URI  types.String `tfsdk:"uri"`
}

func eventSourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"uri":  types.StringType,
	}
}

func expandEventSources(ctx context.Context, list types.List, diags *diag.Diagnostics) []ngrok.EventSourceReplace {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var models []eventSourceModel
	diags.Append(list.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil
	}

	sources := make([]ngrok.EventSourceReplace, len(models))
	for i, m := range models {
		sources[i] = ngrok.EventSourceReplace{
			Type: m.Type.ValueString(),
		}
	}
	return sources
}

func flattenEventSources(ctx context.Context, sources []ngrok.EventSource, diags *diag.Diagnostics) types.List {
	if sources == nil {
		return types.ListNull(types.ObjectType{AttrTypes: eventSourceAttrTypes()})
	}

	models := make([]eventSourceModel, len(sources))
	for i, s := range sources {
		models[i] = eventSourceModel{
			Type: types.StringValue(s.Type),
			URI:  types.StringValue(s.URI),
		}
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: eventSourceAttrTypes()}, models)
	diags.Append(d...)
	return list
}

func flattenEventSubscription(ctx context.Context, sub *ngrok.EventSubscription, model *eventSubscriptionResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(sub.ID)
	model.Description = types.StringValue(sub.Description)
	model.Metadata = types.StringValue(sub.Metadata)
	model.URI = types.StringValue(sub.URI)
	model.CreatedAt = types.StringValue(sub.CreatedAt)
	model.Sources = flattenEventSources(ctx, sub.Sources, diags)

	destIDs, d := types.ListValueFrom(ctx, types.StringType, flattenRefList(sub.Destinations))
	diags.Append(d...)
	model.DestinationIDs = destIDs
}
