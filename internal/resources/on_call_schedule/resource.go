package on_call_schedule

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &onCallScheduleResource{}
	_ resource.ResourceWithImportState = &onCallScheduleResource{}
)

type onCallScheduleResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &onCallScheduleResource{}
}

func (r *onCallScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_schedule"
}

func (r *onCallScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift on-call schedule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the on-call schedule.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the on-call schedule (max 255 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"scope_type": schema.StringAttribute{
				Description: "Schedule scope: org_default (organization-wide) or monitor_group (scoped to a specific group).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("org_default"),
				Validators: []validator.String{
					stringvalidator.OneOf("org_default", "monitor_group"),
				},
			},
			"monitor_group_id": schema.Int64Attribute{
				Description: "The monitor group ID this schedule is scoped to. Required when scope_type is monitor_group.",
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "IANA timezone for the schedule (e.g. America/New_York, UTC).",
				Required:    true,
			},
			"frequency": schema.StringAttribute{
				Description: "Rotation frequency: daily or weekly.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("daily", "weekly"),
				},
			},
			"rotation_interval": schema.Int64Attribute{
				Description: "Number of frequency units per rotation (e.g. 1 = every day/week, 2 = every other).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"starts_at": schema.StringAttribute{
				Description: "When the schedule becomes active, in ISO 8601 format.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the on-call schedule is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"participant_ids": schema.ListAttribute{
				Description: "Ordered list of organization member IDs in the rotation. Order determines rotation position.",
				Required:    true,
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (r *onCallScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *onCallScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OnCallScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := modelToInput(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateOnCallSchedule(input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating on-call schedule", fmt.Sprintf("Could not create on-call schedule: %s", err))
		return
	}

	apiToModel(result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created on-call schedule", map[string]any{"id": result.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *onCallScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OnCallScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	result, err := r.client.GetOnCallSchedule(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "on-call schedule not found, removing from state", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading on-call schedule", fmt.Sprintf("Could not read on-call schedule %d: %s", id, err))
		return
	}

	apiToModel(result, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *onCallScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OnCallScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state OnCallScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	input := modelToInput(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateOnCallSchedule(id, input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating on-call schedule", fmt.Sprintf("Could not update on-call schedule %d: %s", id, err))
		return
	}

	apiToModel(result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated on-call schedule", map[string]any{"id": result.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *onCallScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OnCallScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	err := r.client.DeleteOnCallSchedule(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "on-call schedule already deleted", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError("Error deleting on-call schedule", fmt.Sprintf("Could not delete on-call schedule %d: %s", id, err))
		return
	}

	tflog.Trace(ctx, "deleted on-call schedule", map[string]any{"id": id})
}

func (r *onCallScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Could not parse on-call schedule ID %q: %s", req.ID, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func modelToInput(ctx context.Context, m *OnCallScheduleModel, diags *diag.Diagnostics) apiclient.OnCallScheduleInput {
	input := apiclient.OnCallScheduleInput{
		Name:             m.Name.ValueString(),
		ScopeType:        m.ScopeType.ValueString(),
		MonitorGroupID:   helpers.IntPtr(m.MonitorGroupID),
		Timezone:         m.Timezone.ValueString(),
		Frequency:        m.Frequency.ValueString(),
		RotationInterval: int(m.RotationInterval.ValueInt64()),
		StartsAt:         m.StartsAt.ValueString(),
		Enabled:          m.Enabled.ValueBool(),
	}

	if !m.ParticipantIDs.IsNull() && !m.ParticipantIDs.IsUnknown() {
		var ids []int64
		diags.Append(m.ParticipantIDs.ElementsAs(ctx, &ids, false)...)
		intIDs := make([]int, len(ids))
		for i, v := range ids {
			intIDs[i] = int(v)
		}
		input.ParticipantIDs = intIDs
	}

	return input
}

func apiToModel(s *apiclient.OnCallSchedule, m *OnCallScheduleModel, diags *diag.Diagnostics) {
	m.ID = types.Int64Value(int64(s.ID))
	m.Name = types.StringValue(s.Name)
	m.ScopeType = types.StringValue(s.ScopeType)
	if s.MonitorGroup != nil {
		m.MonitorGroupID = types.Int64Value(int64(s.MonitorGroup.ID))
	} else {
		m.MonitorGroupID = types.Int64Null()
	}
	m.Timezone = types.StringValue(s.Timezone)
	m.Frequency = types.StringValue(s.Frequency)
	m.RotationInterval = types.Int64Value(int64(s.RotationInterval))
	// Normalize +00:00 → Z to avoid spurious diffs when the API returns
	// a different but semantically equivalent ISO 8601 offset.
	m.StartsAt = types.StringValue(normalizeTimestamp(s.StartsAt))
	m.Enabled = types.BoolValue(s.Enabled)

	// Sort participants by position to maintain stable ordering.
	sorted := make([]apiclient.OnCallParticipant, len(s.Participants))
	copy(sorted, s.Participants)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Position < sorted[j].Position
	})

	if len(sorted) > 0 {
		ids := make([]attr.Value, len(sorted))
		for i, p := range sorted {
			ids[i] = types.Int64Value(int64(p.OrganizationMember.ID))
		}
		listVal, d := types.ListValue(types.Int64Type, ids)
		diags.Append(d...)
		m.ParticipantIDs = listVal
	} else {
		m.ParticipantIDs = types.ListNull(types.Int64Type)
	}
}

// normalizeTimestamp converts "+00:00" suffix to "Z" for consistent comparison.
func normalizeTimestamp(ts string) string {
	return strings.Replace(ts, "+00:00", "Z", 1)
}
