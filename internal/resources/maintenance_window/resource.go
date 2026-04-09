package maintenance_window

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &maintenanceWindowResource{}
	_ resource.ResourceWithImportState = &maintenanceWindowResource{}
)

type maintenanceWindowResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &maintenanceWindowResource{}
}

func (r *maintenanceWindowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_window"
}

func (r *maintenanceWindowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift maintenance window.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the maintenance window.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the maintenance window.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"repeat": schema.StringAttribute{
				Description: "The repeat schedule: none, daily, weekly, or monthly.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("none", "daily", "weekly", "monthly"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the maintenance window is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"target_type": schema.StringAttribute{
				Description: "The type of target: all, monitor, group, or tag.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("all", "monitor", "group", "tag"),
				},
			},
			"target_id": schema.Int64Attribute{
				Description: "The ID of the target. Required when target_type is not 'all'.",
				Optional:    true,
			},
			"start_date_time": schema.StringAttribute{
				Description: "The start date and time in ISO 8601 format. Required when repeat is 'none'.",
				Optional:    true,
			},
			"end_date_time": schema.StringAttribute{
				Description: "The end date and time in ISO 8601 format. Required when repeat is 'none'.",
				Optional:    true,
			},
			"start_time": schema.StringAttribute{
				Description: "The start time in HH:MM format. Required when repeat is not 'none'.",
				Optional:    true,
			},
			"end_time": schema.StringAttribute{
				Description: "The end time in HH:MM format. Required when repeat is not 'none'.",
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "The timezone for the maintenance window.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("UTC"),
			},
			"weekly_days": schema.ListAttribute{
				Description: "List of weekday names for weekly repeat schedules.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"monthly_day_type": schema.StringAttribute{
				Description: "The type of monthly day: first, last, or specific.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("first", "last", "specific"),
				},
			},
			"monthly_day": schema.Int64Attribute{
				Description: "The day of the month (1-31). Used when monthly_day_type is 'specific'.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 31),
				},
			},
		},
	}
}

func (r *maintenanceWindowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *maintenanceWindowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MaintenanceWindowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := apiclient.MaintenanceWindowInput{
		Name:       plan.Name.ValueString(),
		Repeat:     plan.Repeat.ValueString(),
		TargetType: plan.TargetType.ValueString(),
		Timezone:   plan.Timezone.ValueString(),
	}

	input.Enabled = helpers.BoolPtr(plan.Enabled)
	input.TargetID = helpers.IntPtr(plan.TargetID)
	input.StartDateTime = helpers.StringPtr(plan.StartDateTime)
	input.EndDateTime = helpers.StringPtr(plan.EndDateTime)
	input.StartTime = helpers.StringPtr(plan.StartTime)
	input.EndTime = helpers.StringPtr(plan.EndTime)

	if !plan.WeeklyDays.IsNull() && !plan.WeeklyDays.IsUnknown() {
		var days []string
		resp.Diagnostics.Append(plan.WeeklyDays.ElementsAs(ctx, &days, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.WeeklyDays = days
	}

	input.MonthlyDayType = helpers.StringPtr(plan.MonthlyDayType)
	input.MonthlyDay = helpers.IntPtr(plan.MonthlyDay)

	mw, err := r.client.CreateMaintenanceWindow(input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating maintenance window", err.Error())
		return
	}

	// Only take the ID from the response; keep plan values for everything
	// else since the API Create response may omit or default fields.
	plan.ID = types.Int64Value(int64(mw.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *maintenanceWindowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MaintenanceWindowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	mw, err := r.client.GetMaintenanceWindow(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading maintenance window", err.Error())
		return
	}

	mapResponseToState(ctx, mw, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *maintenanceWindowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MaintenanceWindowModel
	var state MaintenanceWindowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	patch := apiclient.MaintenanceWindowPatch{}

	if !plan.Name.Equal(state.Name) {
		patch["name"] = plan.Name.ValueString()
	}

	if !plan.Repeat.Equal(state.Repeat) {
		patch["repeat"] = plan.Repeat.ValueString()
	}

	if !plan.Enabled.Equal(state.Enabled) {
		patch["enabled"] = plan.Enabled.ValueBool()
	}

	if !plan.TargetType.Equal(state.TargetType) {
		patch["targetType"] = plan.TargetType.ValueString()
	}

	if !plan.TargetID.Equal(state.TargetID) {
		if plan.TargetID.IsNull() {
			patch["targetId"] = nil
		} else {
			patch["targetId"] = int(plan.TargetID.ValueInt64())
		}
	}

	if !plan.StartDateTime.Equal(state.StartDateTime) {
		if plan.StartDateTime.IsNull() {
			patch["startDateTime"] = nil
		} else {
			patch["startDateTime"] = plan.StartDateTime.ValueString()
		}
	}

	if !plan.EndDateTime.Equal(state.EndDateTime) {
		if plan.EndDateTime.IsNull() {
			patch["endDateTime"] = nil
		} else {
			patch["endDateTime"] = plan.EndDateTime.ValueString()
		}
	}

	if !plan.StartTime.Equal(state.StartTime) {
		if plan.StartTime.IsNull() {
			patch["startTime"] = nil
		} else {
			patch["startTime"] = plan.StartTime.ValueString()
		}
	}

	if !plan.EndTime.Equal(state.EndTime) {
		if plan.EndTime.IsNull() {
			patch["endTime"] = nil
		} else {
			patch["endTime"] = plan.EndTime.ValueString()
		}
	}

	if !plan.Timezone.Equal(state.Timezone) {
		patch["timezone"] = plan.Timezone.ValueString()
	}

	if !plan.WeeklyDays.Equal(state.WeeklyDays) {
		if plan.WeeklyDays.IsNull() {
			patch["weeklyDays"] = nil
		} else {
			var days []string
			resp.Diagnostics.Append(plan.WeeklyDays.ElementsAs(ctx, &days, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			patch["weeklyDays"] = days
		}
	}

	if !plan.MonthlyDayType.Equal(state.MonthlyDayType) {
		if plan.MonthlyDayType.IsNull() {
			patch["monthlyDayType"] = nil
		} else {
			patch["monthlyDayType"] = plan.MonthlyDayType.ValueString()
		}
	}

	if !plan.MonthlyDay.Equal(state.MonthlyDay) {
		if plan.MonthlyDay.IsNull() {
			patch["monthlyDay"] = nil
		} else {
			patch["monthlyDay"] = int(plan.MonthlyDay.ValueInt64())
		}
	}

	_, err := r.client.UpdateMaintenanceWindow(id, patch)
	if err != nil {
		resp.Diagnostics.AddError("Error updating maintenance window", err.Error())
		return
	}

	// Keep plan values — the PATCH response may omit fields.
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *maintenanceWindowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MaintenanceWindowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	err := r.client.DeleteMaintenanceWindow(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting maintenance window", err.Error())
	}
}

func (r *maintenanceWindowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Could not parse ID %q as integer: %s", req.ID, err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(id))...)
}

func mapResponseToState(ctx context.Context, mw *apiclient.MaintenanceWindow, state *MaintenanceWindowModel, diags *diag.Diagnostics) {
	state.ID = types.Int64Value(int64(mw.ID))
	state.Name = types.StringValue(mw.Name)
	state.Repeat = types.StringValue(mw.Repeat)
	state.Enabled = types.BoolValue(mw.Enabled)
	state.TargetType = types.StringValue(mw.TargetType)
	state.Timezone = types.StringValue(mw.Timezone)

	if mw.TargetID != nil {
		state.TargetID = types.Int64Value(int64(*mw.TargetID))
	} else if state.TargetID.IsNull() {
		// keep null
	}
	// else: API returned nil but state has a value — preserve it

	if mw.StartDateTime != nil {
		state.StartDateTime = types.StringValue(normalizeTimestamp(*mw.StartDateTime, state.StartDateTime))
	} else {
		state.StartDateTime = types.StringNull()
	}

	if mw.EndDateTime != nil {
		state.EndDateTime = types.StringValue(normalizeTimestamp(*mw.EndDateTime, state.EndDateTime))
	} else {
		state.EndDateTime = types.StringNull()
	}

	if mw.StartTime != nil {
		state.StartTime = types.StringValue(*mw.StartTime)
	} else {
		state.StartTime = types.StringNull()
	}

	if mw.EndTime != nil {
		state.EndTime = types.StringValue(*mw.EndTime)
	} else {
		state.EndTime = types.StringNull()
	}

	if len(mw.WeeklyDays) > 0 {
		weeklyDays, d := types.ListValueFrom(ctx, types.StringType, mw.WeeklyDays)
		diags.Append(d...)
		state.WeeklyDays = weeklyDays
	} else {
		state.WeeklyDays = types.ListNull(types.StringType)
	}

	if mw.MonthlyDayType != nil {
		state.MonthlyDayType = types.StringValue(*mw.MonthlyDayType)
	} else {
		state.MonthlyDayType = types.StringNull()
	}

	if mw.MonthlyDay != nil {
		state.MonthlyDay = types.Int64Value(int64(*mw.MonthlyDay))
	} else {
		state.MonthlyDay = types.Int64Null()
	}
}

// normalizeTimestamp keeps the state/plan timestamp format when the API returns
// a semantically equivalent value (e.g. "Z" vs "+00:00").
func normalizeTimestamp(apiVal string, stateVal types.String) string {
	if stateVal.IsNull() || stateVal.IsUnknown() {
		return apiVal
	}
	apiTime, err1 := time.Parse(time.RFC3339, apiVal)
	stateTime, err2 := time.Parse(time.RFC3339, stateVal.ValueString())
	if err1 == nil && err2 == nil && apiTime.Equal(stateTime) {
		return stateVal.ValueString()
	}
	return apiVal
}
