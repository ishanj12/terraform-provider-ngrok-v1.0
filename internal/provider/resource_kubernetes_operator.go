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
	"github.com/ngrok/ngrok-api-go/v9/kubernetes_operators"
)

var (
	_ resource.Resource                   = &kubernetesOperatorResource{}
	_ resource.ResourceWithImportState    = &kubernetesOperatorResource{}
	_ resource.ResourceWithModifyPlan     = &kubernetesOperatorResource{}
)

type kubernetesOperatorResourceModel struct {
	ID              types.String   `tfsdk:"id"`
	URI             types.String   `tfsdk:"uri"`
	CreatedAt       types.String   `tfsdk:"created_at"`
	UpdatedAt       types.String   `tfsdk:"updated_at"`
	Description     types.String   `tfsdk:"description"`
	Metadata        types.String   `tfsdk:"metadata"`
	EnabledFeatures []types.String `tfsdk:"enabled_features"`
	Region          types.String   `tfsdk:"region"`
	Deployment      types.Object   `tfsdk:"deployment"`
	Binding         types.Object   `tfsdk:"binding"`
	PrincipalID     types.String   `tfsdk:"principal_id"`
}

type kubernetesOperatorResource struct {
	client *kubernetes_operators.Client
}

func NewKubernetesOperatorResource() resource.Resource {
	return &kubernetesOperatorResource{}
}

func (r *kubernetesOperatorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_operator"
}

func (r *kubernetesOperatorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Kubernetes Operator registered with ngrok. Used by the Kubernetes Operator to register and manage its own resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this Kubernetes Operator.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uri": schema.StringAttribute{
				Description: "URI of this Kubernetes Operator API resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the Kubernetes Operator was created, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the Kubernetes Operator was last updated, RFC 3339 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of this Kubernetes Operator. Optional, max 255 bytes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Arbitrary user-defined machine-readable data of this Kubernetes Operator. Optional, max 4096 bytes.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled_features": schema.ListAttribute{
				Description: `Features enabled for this Kubernetes Operator. A subset of "bindings", "ingress", and "gateway".`,
				Optional:    true,
				ElementType: types.StringType,
			},
			"region": schema.StringAttribute{
				Description: `The ngrok region in which the ingress for this operator is served. Defaults to "global".`,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment": schema.SingleNestedAttribute{
				Description: "Information about the deployment of this Kubernetes Operator.",
				Required:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "The deployment name.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"namespace": schema.StringAttribute{
						Description: "The namespace this Kubernetes Operator is deployed to.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"version": schema.StringAttribute{
						Description: "The version of this Kubernetes Operator.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"cluster_name": schema.StringAttribute{
						Description: "User-given name for the cluster the Kubernetes Operator is deployed to.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"binding": schema.SingleNestedAttribute{
				Description: "Configuration for the Bindings feature of this Kubernetes Operator.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"endpoint_selectors": schema.ListAttribute{
						Description: "The list of CEL expressions that filter the k8s bound endpoints for this operator.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"csr": schema.StringAttribute{
						Description: "CSR supplied during initial creation to enable mutual TLS between ngrok and the operator.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"ingress_endpoint": schema.StringAttribute{
						Description: "The public ingress endpoint for this Kubernetes Operator.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"cert": schema.SingleNestedAttribute{
						Description: "The binding certificate information.",
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"cert": schema.StringAttribute{
								Description: "The public client certificate generated from the CSR.",
								Computed:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"not_before": schema.StringAttribute{
								Description: "Timestamp when the certificate becomes valid, RFC 3339 format.",
								Computed:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"not_after": schema.StringAttribute{
								Description: "Timestamp when the certificate becomes invalid, RFC 3339 format.",
								Computed:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
			"principal_id": schema.StringAttribute{
				Description: "The ID of the principal who created this Kubernetes Operator.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *kubernetesOperatorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = kubernetes_operators.NewClient(clientConfig)
}

func (r *kubernetesOperatorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan kubernetesOperatorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &ngrok.KubernetesOperatorCreate{
		Description:     plan.Description.ValueString(),
		Metadata:        plan.Metadata.ValueString(),
		EnabledFeatures: expandStringList(plan.EnabledFeatures),
		Region:          plan.Region.ValueString(),
		Deployment:      expandK8sOperatorDeployment(ctx, plan.Deployment, &resp.Diagnostics),
		Binding:         expandK8sOperatorBindingCreate(ctx, plan.Binding, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	operator, err := r.client.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating kubernetes operator", err.Error())
		return
	}

	flattenKubernetesOperator(ctx, operator, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kubernetesOperatorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state kubernetesOperatorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	operator, err := r.client.Get(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading kubernetes operator", err.Error())
		return
	}

	flattenKubernetesOperator(ctx, operator, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *kubernetesOperatorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan kubernetesOperatorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state kubernetesOperatorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the raw config to check if binding was user-configured.
	// UseStateForUnknown copies binding from state into plan even when
	// the user didn't set it, so we must check config to avoid sending
	// binding to the API when the user didn't configure it.
	var config kubernetesOperatorResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &ngrok.KubernetesOperatorUpdate{
		ID:              state.ID.ValueString(),
		Description:     stringPtrFromFramework(plan.Description),
		Metadata:        stringPtrFromFramework(plan.Metadata),
		EnabledFeatures: expandStringList(plan.EnabledFeatures),
		Region:          stringPtrFromFramework(plan.Region),
	}
	// Only send binding if the user explicitly configured it
	if !config.Binding.IsNull() {
		updateReq.Binding = expandK8sOperatorBindingUpdate(ctx, plan.Binding, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	operator, err := r.client.Update(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating kubernetes operator", err.Error())
		return
	}

	flattenKubernetesOperator(ctx, operator, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kubernetesOperatorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state kubernetesOperatorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting kubernetes operator", err.Error())
	}
}

func (r *kubernetesOperatorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *kubernetesOperatorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on create or destroy
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan kubernetesOperatorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state kubernetesOperatorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If any user-configurable field changed, mark updated_at as unknown
	// so the provider can return the new value from the API.
	hasChanges := !plan.Description.Equal(state.Description) ||
		!plan.Metadata.Equal(state.Metadata) ||
		!plan.Region.Equal(state.Region) ||
		!plan.Deployment.Equal(state.Deployment) ||
		!plan.Binding.Equal(state.Binding) ||
		!stringListEqual(plan.EnabledFeatures, state.EnabledFeatures)

	if hasChanges {
		resp.Plan.SetAttribute(ctx, path.Root("updated_at"), types.StringUnknown())
	}
}

// --- Flatten ---

func flattenKubernetesOperator(ctx context.Context, op *ngrok.KubernetesOperator, model *kubernetesOperatorResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(op.ID)
	model.URI = types.StringValue(op.URI)
	model.CreatedAt = types.StringValue(op.CreatedAt)
	model.UpdatedAt = types.StringValue(op.UpdatedAt)
	model.Description = types.StringValue(op.Description)
	model.Metadata = types.StringValue(op.Metadata)
	model.EnabledFeatures = flattenStringList(op.EnabledFeatures)
	model.Region = types.StringValue(op.Region)
	model.PrincipalID = types.StringValue(op.Principal.ID)
	// Merge deployment fields: preserve plan/state values for fields the API returns as empty,
	// but always populate from API on import (when model.Deployment is null).
	if !model.Deployment.IsNull() && !model.Deployment.IsUnknown() {
		// Merge: use API values when non-empty, preserve model values otherwise
		var planDep k8sOperatorDeploymentModel
		diags.Append(model.Deployment.As(ctx, &planDep, basetypes.ObjectAsOptions{})...)
		merged := k8sOperatorDeploymentModel{
			Name:        planDep.Name,
			Namespace:   planDep.Namespace,
			Version:     planDep.Version,
			ClusterName: planDep.ClusterName,
		}
		if op.Deployment.Name != "" {
			merged.Name = types.StringValue(op.Deployment.Name)
		}
		if op.Deployment.Namespace != "" {
			merged.Namespace = types.StringValue(op.Deployment.Namespace)
		}
		if op.Deployment.Version != "" {
			merged.Version = types.StringValue(op.Deployment.Version)
		}
		if op.Deployment.ClusterName != "" {
			merged.ClusterName = types.StringValue(op.Deployment.ClusterName)
		}
		obj, d := types.ObjectValueFrom(ctx, k8sOperatorDeploymentAttrTypes(), &merged)
		diags.Append(d...)
		model.Deployment = obj
	} else {
		model.Deployment = flattenK8sOperatorDeployment(ctx, &op.Deployment, diags)
	}
	// Flatten binding from API. On import or when user didn't configure it,
	// still populate from the API so state reflects reality.
	if !model.Binding.IsNull() && !model.Binding.IsUnknown() {
		model.Binding = flattenK8sOperatorBinding(ctx, op.Binding, model.Binding, diags)
	} else {
		model.Binding = flattenK8sOperatorBindingFresh(ctx, op.Binding, diags)
	}
}

// --- Deployment ---

type k8sOperatorDeploymentModel struct {
	Name        types.String `tfsdk:"name"`
	Namespace   types.String `tfsdk:"namespace"`
	Version     types.String `tfsdk:"version"`
	ClusterName types.String `tfsdk:"cluster_name"`
}

func k8sOperatorDeploymentAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":         types.StringType,
		"namespace":    types.StringType,
		"version":      types.StringType,
		"cluster_name": types.StringType,
	}
}

func expandK8sOperatorDeployment(ctx context.Context, obj types.Object, diags *diag.Diagnostics) ngrok.KubernetesOperatorDeployment {
	if obj.IsNull() || obj.IsUnknown() {
		return ngrok.KubernetesOperatorDeployment{}
	}
	var model k8sOperatorDeploymentModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return ngrok.KubernetesOperatorDeployment{}
	}
	return ngrok.KubernetesOperatorDeployment{
		Name:        model.Name.ValueString(),
		Namespace:   model.Namespace.ValueString(),
		Version:     model.Version.ValueString(),
		ClusterName: model.ClusterName.ValueString(),
	}
}

func flattenK8sOperatorDeployment(ctx context.Context, dep *ngrok.KubernetesOperatorDeployment, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValueFrom(ctx, k8sOperatorDeploymentAttrTypes(), &k8sOperatorDeploymentModel{
		Name:        types.StringValue(dep.Name),
		Namespace:   types.StringValue(dep.Namespace),
		Version:     types.StringValue(dep.Version),
		ClusterName: types.StringValue(dep.ClusterName),
	})
	diags.Append(d...)
	return obj
}

// --- Binding ---

type k8sOperatorBindingModel struct {
	EndpointSelectors []types.String `tfsdk:"endpoint_selectors"`
	CSR               types.String   `tfsdk:"csr"`
	IngressEndpoint   types.String   `tfsdk:"ingress_endpoint"`
	Cert              types.Object   `tfsdk:"cert"`
}

type k8sOperatorCertModel struct {
	Cert      types.String `tfsdk:"cert"`
	NotBefore types.String `tfsdk:"not_before"`
	NotAfter  types.String `tfsdk:"not_after"`
}

func k8sOperatorCertAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cert":       types.StringType,
		"not_before": types.StringType,
		"not_after":  types.StringType,
	}
}

func k8sOperatorBindingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"endpoint_selectors": types.ListType{ElemType: types.StringType},
		"csr":                types.StringType,
		"ingress_endpoint":   types.StringType,
		"cert":               types.ObjectType{AttrTypes: k8sOperatorCertAttrTypes()},
	}
}

func expandK8sOperatorBindingCreate(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *ngrok.KubernetesOperatorBindingCreate {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model k8sOperatorBindingModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return &ngrok.KubernetesOperatorBindingCreate{
		EndpointSelectors: expandStringList(model.EndpointSelectors),
		CSR:               model.CSR.ValueString(),
		IngressEndpoint:   stringPtrFromFramework(model.IngressEndpoint),
	}
}

func expandK8sOperatorBindingUpdate(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *ngrok.KubernetesOperatorBindingUpdate {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model k8sOperatorBindingModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return &ngrok.KubernetesOperatorBindingUpdate{
		EndpointSelectors: expandStringList(model.EndpointSelectors),
		IngressEndpoint:   stringPtrFromFramework(model.IngressEndpoint),
	}
}

func flattenK8sOperatorBinding(ctx context.Context, binding *ngrok.KubernetesOperatorBinding, priorBinding types.Object, diags *diag.Diagnostics) types.Object {
	if binding == nil {
		return types.ObjectNull(k8sOperatorBindingAttrTypes())
	}

	certObj, d := types.ObjectValueFrom(ctx, k8sOperatorCertAttrTypes(), &k8sOperatorCertModel{
		Cert:      types.StringValue(binding.Cert.Cert),
		NotBefore: types.StringValue(binding.Cert.NotBefore),
		NotAfter:  types.StringValue(binding.Cert.NotAfter),
	})
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(k8sOperatorBindingAttrTypes())
	}

	// Handle endpoint_selectors: preserve null from plan/state when API returns empty list
	var selectors basetypes.ListValue
	if len(binding.EndpointSelectors) > 0 {
		selectors, d = types.ListValueFrom(ctx, types.StringType, flattenStringList(binding.EndpointSelectors))
		diags.Append(d...)
	} else {
		// Check if the prior state had endpoint_selectors set
		var priorModel k8sOperatorBindingModel
		diags.Append(priorBinding.As(ctx, &priorModel, basetypes.ObjectAsOptions{})...)
		if priorModel.EndpointSelectors != nil {
			selectors, d = types.ListValueFrom(ctx, types.StringType, []types.String{})
			diags.Append(d...)
		} else {
			selectors = types.ListNull(types.StringType)
		}
	}
	if diags.HasError() {
		return types.ObjectNull(k8sOperatorBindingAttrTypes())
	}

	// Preserve CSR from prior state since the API doesn't return it
	csrValue := types.StringValue("")
	if !priorBinding.IsNull() && !priorBinding.IsUnknown() {
		var priorModel k8sOperatorBindingModel
		diags.Append(priorBinding.As(ctx, &priorModel, basetypes.ObjectAsOptions{})...)
		if !priorModel.CSR.IsNull() && !priorModel.CSR.IsUnknown() {
			csrValue = priorModel.CSR
		}
	}

	obj, d := types.ObjectValue(k8sOperatorBindingAttrTypes(), map[string]attr.Value{
		"endpoint_selectors": selectors,
		"csr":                csrValue,
		"ingress_endpoint":   types.StringValue(binding.IngressEndpoint),
		"cert":               certObj,
	})
	diags.Append(d...)
	return obj
}

// flattenK8sOperatorBindingFresh flattens a binding from the API without any
// prior state context. Used on import where no prior state exists.
func flattenK8sOperatorBindingFresh(ctx context.Context, binding *ngrok.KubernetesOperatorBinding, diags *diag.Diagnostics) types.Object {
	if binding == nil {
		return types.ObjectNull(k8sOperatorBindingAttrTypes())
	}

	certObj, d := types.ObjectValueFrom(ctx, k8sOperatorCertAttrTypes(), &k8sOperatorCertModel{
		Cert:      types.StringValue(binding.Cert.Cert),
		NotBefore: types.StringValue(binding.Cert.NotBefore),
		NotAfter:  types.StringValue(binding.Cert.NotAfter),
	})
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(k8sOperatorBindingAttrTypes())
	}

	var selectors basetypes.ListValue
	if len(binding.EndpointSelectors) > 0 {
		selectors, d = types.ListValueFrom(ctx, types.StringType, flattenStringList(binding.EndpointSelectors))
		diags.Append(d...)
	} else {
		selectors = types.ListNull(types.StringType)
	}
	if diags.HasError() {
		return types.ObjectNull(k8sOperatorBindingAttrTypes())
	}

	obj, d := types.ObjectValue(k8sOperatorBindingAttrTypes(), map[string]attr.Value{
		"endpoint_selectors": selectors,
		"csr":                types.StringValue(""),
		"ingress_endpoint":   types.StringValue(binding.IngressEndpoint),
		"cert":               certObj,
	})
	diags.Append(d...)
	return obj
}
