package on_call_schedule_override

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &onCallScheduleOverrideResource{}
	_ resource.ResourceWithImportState = &onCallScheduleOverrideResource{}
)

type onCallScheduleOverrideResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &onCallScheduleOverrideResource{}
}

func (r *onCallScheduleOverrideResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_schedule_override"
}

func (r *onCallScheduleOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift on-call schedule override.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the override.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"schedule_id": schema.Int64Attribute{
				Description: "The ID of the parent on-call schedule.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"replacement_member_id": schema.Int64Attribute{
				Description: "The organization member ID who will cover during this override.",
				Required:    true,
			},
			"starts_at": schema.StringAttribute{
				Description: "Override start time in ISO 8601 format.",
				Required:    true,
			},
			"ends_at": schema.StringAttribute{
				Description: "Override end time in ISO 8601 format.",
				Required:    true,
			},
			"note": schema.StringAttribute{
				Description: "Optional note for the override (max 255 characters).",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
		},
	}
}

func (r *onCallScheduleOverrideResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *onCallScheduleOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OnCallScheduleOverrideModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := int(plan.ScheduleID.ValueInt64())
	input := apiclient.OnCallOverrideInput{
		ReplacementMemberID: int(plan.ReplacementMemberID.ValueInt64()),
		StartsAt:            plan.StartsAt.ValueString(),
		EndsAt:              plan.EndsAt.ValueString(),
		Note:                helpers.StringPtr(plan.Note),
	}

	result, err := r.client.CreateOnCallOverride(scheduleID, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating on-call override", fmt.Sprintf("Could not create on-call override: %s", err))
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.ReplacementMemberID = types.Int64Value(int64(result.ReplacementMember.ID))
	plan.StartsAt = types.StringValue(normalizeTimestamp(result.StartsAt))
	plan.EndsAt = types.StringValue(normalizeTimestamp(result.EndsAt))
	if result.Note != nil {
		plan.Note = types.StringValue(*result.Note)
	} else {
		plan.Note = types.StringNull()
	}

	tflog.Trace(ctx, "created on-call override", map[string]any{"id": result.ID, "schedule_id": scheduleID})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *onCallScheduleOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OnCallScheduleOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := int(state.ScheduleID.ValueInt64())
	overrideID := int(state.ID.ValueInt64())

	result, err := r.client.GetOnCallOverride(scheduleID, overrideID)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "on-call override not found, removing from state", map[string]any{"id": overrideID, "schedule_id": scheduleID})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading on-call override", fmt.Sprintf("Could not read on-call override %d: %s", overrideID, err))
		return
	}

	state.ID = types.Int64Value(int64(result.ID))
	state.ReplacementMemberID = types.Int64Value(int64(result.ReplacementMember.ID))
	state.StartsAt = types.StringValue(normalizeTimestamp(result.StartsAt))
	state.EndsAt = types.StringValue(normalizeTimestamp(result.EndsAt))
	if result.Note != nil {
		state.Note = types.StringValue(*result.Note)
	} else {
		state.Note = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *onCallScheduleOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OnCallScheduleOverrideModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state OnCallScheduleOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := int(state.ScheduleID.ValueInt64())
	overrideID := int(state.ID.ValueInt64())

	input := apiclient.OnCallOverrideInput{
		ReplacementMemberID: int(plan.ReplacementMemberID.ValueInt64()),
		StartsAt:            plan.StartsAt.ValueString(),
		EndsAt:              plan.EndsAt.ValueString(),
		Note:                helpers.StringPtr(plan.Note),
	}

	result, err := r.client.UpdateOnCallOverride(scheduleID, overrideID, input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating on-call override", fmt.Sprintf("Could not update on-call override %d: %s", overrideID, err))
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.ReplacementMemberID = types.Int64Value(int64(result.ReplacementMember.ID))
	plan.StartsAt = types.StringValue(normalizeTimestamp(result.StartsAt))
	plan.EndsAt = types.StringValue(normalizeTimestamp(result.EndsAt))
	if result.Note != nil {
		plan.Note = types.StringValue(*result.Note)
	} else {
		plan.Note = types.StringNull()
	}

	tflog.Trace(ctx, "updated on-call override", map[string]any{"id": result.ID, "schedule_id": scheduleID})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *onCallScheduleOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OnCallScheduleOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID := int(state.ScheduleID.ValueInt64())
	overrideID := int(state.ID.ValueInt64())

	err := r.client.DeleteOnCallOverride(scheduleID, overrideID)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "on-call override already deleted", map[string]any{"id": overrideID})
			return
		}
		resp.Diagnostics.AddError("Error deleting on-call override", fmt.Sprintf("Could not delete on-call override %d: %s", overrideID, err))
		return
	}

	tflog.Trace(ctx, "deleted on-call override", map[string]any{"id": overrideID, "schedule_id": scheduleID})
}

func (r *onCallScheduleOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Expected import ID in format '{schedule_id}/{override_id}', got: %q", req.ID))
		return
	}

	scheduleID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Could not parse schedule_id %q as integer: %s", parts[0], err))
		return
	}

	overrideID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Could not parse override_id %q as integer: %s", parts[1], err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schedule_id"), scheduleID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), overrideID)...)
}

func normalizeTimestamp(ts string) string {
	return strings.Replace(ts, "+00:00", "Z", 1)
}
