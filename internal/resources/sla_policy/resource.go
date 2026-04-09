package sla_policy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &slaPolicyResource{}
	_ resource.ResourceWithImportState = &slaPolicyResource{}
)

type slaPolicyResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &slaPolicyResource{}
}

func (r *slaPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sla_policy"
}

func (r *slaPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift SLA policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the SLA policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the SLA policy (max 255 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"uptime_target": schema.Float64Attribute{
				Description: "Target uptime percentage (90.0 to 100.0).",
				Required:    true,
				Validators: []validator.Float64{
					float64validator.Between(90.0, 100.0),
				},
			},
			"window_type": schema.StringAttribute{
				Description: "SLA evaluation window: calendar_week, calendar_month, calendar_year, or rolling_30d.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("calendar_week", "calendar_month", "calendar_year", "rolling_30d"),
				},
			},
			"response_time_sla_enabled": schema.BoolAttribute{
				Description: "Whether response time SLA tracking is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"response_time_threshold_ms": schema.Int64Attribute{
				Description: "Response time threshold in milliseconds (1 to 60000). Required when response_time_sla_enabled is true.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 60000),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the SLA policy is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"monitor_ids": schema.ListAttribute{
				Description: "List of monitor IDs directly covered by this policy (max 50).",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"tag_ids": schema.ListAttribute{
				Description: "List of tag IDs for dynamic monitor selection (max 20).",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"group_ids": schema.ListAttribute{
				Description: "List of monitor group IDs covered by this policy (max 20).",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"created_at": schema.StringAttribute{
				Description: "The date/time the SLA policy was created, in ISO 8601 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The date/time the SLA policy was last updated, in ISO 8601 format.",
				Computed:    true,
			},
		},
	}
}

func (r *slaPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *slaPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SlaPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := apiclient.SlaPolicyCreateInput{
		Name:         plan.Name.ValueString(),
		UptimeTarget: plan.UptimeTarget.ValueFloat64(),
		WindowType:   plan.WindowType.ValueString(),
	}

	input.ResponseTimeSlaEnabled = helpers.BoolPtr(plan.ResponseTimeSlaEnabled)
	input.ResponseTimeThresholdMs = helpers.IntPtr(plan.ResponseTimeThresholdMs)
	input.Enabled = helpers.BoolPtr(plan.Enabled)

	input.MonitorIDs = listToIntSlice(ctx, plan.MonitorIDs, &resp.Diagnostics)
	input.TagIDs = listToIntSlice(ctx, plan.TagIDs, &resp.Diagnostics)
	input.GroupIDs = listToIntSlice(ctx, plan.GroupIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateSlaPolicy(input)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SLA policy",
			fmt.Sprintf("Could not create SLA policy: %s", err),
		)
		return
	}

	apiToModel(result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created SLA policy", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *slaPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SlaPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	result, err := r.client.GetSlaPolicy(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "SLA policy not found, removing from state", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading SLA policy",
			fmt.Sprintf("Could not read SLA policy %d: %s", id, err),
		)
		return
	}

	apiToModel(result, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *slaPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SlaPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SlaPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	name := plan.Name.ValueString()
	uptimeTarget := plan.UptimeTarget.ValueFloat64()
	windowType := plan.WindowType.ValueString()
	responseTimeSlaEnabled := plan.ResponseTimeSlaEnabled.ValueBool()
	enabled := plan.Enabled.ValueBool()

	input := apiclient.SlaPolicyUpdateInput{
		Name:                   &name,
		UptimeTarget:           &uptimeTarget,
		WindowType:             &windowType,
		ResponseTimeSlaEnabled: &responseTimeSlaEnabled,
		Enabled:                &enabled,
	}

	input.ResponseTimeThresholdMs = helpers.IntPtr(plan.ResponseTimeThresholdMs)

	input.MonitorIDs = listToIntSlice(ctx, plan.MonitorIDs, &resp.Diagnostics)
	input.TagIDs = listToIntSlice(ctx, plan.TagIDs, &resp.Diagnostics)
	input.GroupIDs = listToIntSlice(ctx, plan.GroupIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateSlaPolicy(id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating SLA policy",
			fmt.Sprintf("Could not update SLA policy %d: %s", id, err),
		)
		return
	}

	apiToModel(result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated SLA policy", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *slaPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SlaPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	err := r.client.DeleteSlaPolicy(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "SLA policy already deleted", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting SLA policy",
			fmt.Sprintf("Could not delete SLA policy %d: %s", id, err),
		)
		return
	}

	tflog.Trace(ctx, "deleted SLA policy", map[string]any{"id": id})
}

func (r *slaPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Could not parse SLA policy ID %q: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// listToIntSlice converts a types.List of Int64 to []int. Returns nil if the list is null.
func listToIntSlice(ctx context.Context, list types.List, diags *diag.Diagnostics) []int {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var ids []int64
	diags.Append(list.ElementsAs(ctx, &ids, false)...)
	if diags.HasError() {
		return nil
	}
	result := make([]int, len(ids))
	for i, v := range ids {
		result[i] = int(v)
	}
	return result
}

// apiToModel maps an API SlaPolicy to the Terraform model.
func apiToModel(p *apiclient.SlaPolicy, m *SlaPolicyModel, diags *diag.Diagnostics) {
	m.ID = types.Int64Value(int64(p.ID))
	m.Name = types.StringValue(p.Name)
	m.UptimeTarget = types.Float64Value(p.UptimeTarget)
	m.WindowType = types.StringValue(p.WindowType)
	m.ResponseTimeSlaEnabled = types.BoolValue(p.ResponseTimeSlaEnabled)
	m.Enabled = types.BoolValue(p.Enabled)
	m.CreatedAt = types.StringValue(p.CreatedAt)
	m.UpdatedAt = types.StringValue(p.UpdatedAt)

	if p.ResponseTimeThresholdMs != nil {
		m.ResponseTimeThresholdMs = types.Int64Value(int64(*p.ResponseTimeThresholdMs))
	} else {
		m.ResponseTimeThresholdMs = types.Int64Null()
	}

	m.MonitorIDs = intsToIDList(p.Monitors, diags)
	m.TagIDs = intsToIDList(p.Tags, diags)
	m.GroupIDs = intsToIDList(p.Groups, diags)
}

type idGetter interface {
	GetID() int
}

// intsToIDList converts a slice of ID-bearing objects to a Terraform types.List of Int64.
func intsToIDList[T idGetter](refs []T, diags *diag.Diagnostics) types.List {
	if len(refs) == 0 {
		return types.ListNull(types.Int64Type)
	}
	ids := make([]attr.Value, len(refs))
	for i, r := range refs {
		ids[i] = types.Int64Value(int64(r.GetID()))
	}
	listVal, d := types.ListValue(types.Int64Type, ids)
	diags.Append(d...)
	return listVal
}
