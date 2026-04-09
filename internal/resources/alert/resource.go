package alert

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var (
	_ resource.Resource                = &alertResource{}
	_ resource.ResourceWithImportState = &alertResource{}
)

type alertResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &alertResource{}
}

func (r *alertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *alertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents a StatusDrift alert. Alerts are read-only and managed through the statusdrift_monitor resource's alert block.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the alert.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"monitor_id": schema.Int64Attribute{
				Description: "The ID of the monitor this alert belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"notification_channel_ids": schema.ListAttribute{
				Description: "List of notification channel IDs to notify when this alert triggers.",
				Required:    true,
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"conditions": schema.ListAttribute{
				Description: "List of conditions that trigger this alert. Valid values: up, down, warning, domain_expiry, cert_expiry, mixed_content, whois_change.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							"up",
							"down",
							"warning",
							"domain_expiry",
							"cert_expiry",
							"mixed_content",
							"whois_change",
						),
					),
				},
			},
			"delay_minutes": schema.Int64Attribute{
				Description: "Number of minutes to delay before sending the alert.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether this alert is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"recurring_interval_minutes": schema.Int64Attribute{
				Description: "Interval in minutes between recurring alert notifications.",
				Optional:    true,
			},
			"max_recurrences": schema.Int64Attribute{
				Description: "Maximum number of recurring notifications to send.",
				Optional:    true,
			},
			"notify_on_call": schema.BoolAttribute{
				Description: "Whether this alert triggers the on-call escalation policy.",
				Computed:    true,
			},
			"on_call_policy_id": schema.Int64Attribute{
				Description: "On-call policy ID used for this alert.",
				Computed:    true,
			},
		},
	}
}

func (r *alertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *alertResource) Create(ctx context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Error(ctx, "alert create is not supported")
	resp.Diagnostics.AddError(
		"Alert Create Not Supported",
		"Alerts cannot be created directly. They are managed through the statusdrift_monitor resource's alert block.",
	)
}

func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	result, err := r.client.GetAlert(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "alert not found, removing from state", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading alert",
			fmt.Sprintf("Could not read alert %d: %s", id, err),
		)
		return
	}

	state.ID = types.Int64Value(int64(result.ID))
	state.MonitorID = types.Int64Value(int64(result.MonitorID))
	state.Enabled = types.BoolValue(result.Enabled)

	// Map notification channel IDs
	channelIDs := make([]types.Int64, len(result.NotificationChannelIDs))
	for i, cid := range result.NotificationChannelIDs {
		channelIDs[i] = types.Int64Value(int64(cid))
	}
	channelIDsList, diags := types.ListValueFrom(ctx, types.Int64Type, channelIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.NotificationChannelIDs = channelIDsList

	// Map conditions
	conditions := make([]types.String, len(result.Conditions))
	for i, c := range result.Conditions {
		conditions[i] = types.StringValue(c)
	}
	conditionsList, diags := types.ListValueFrom(ctx, types.StringType, conditions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Conditions = conditionsList

	// Map optional integer pointer fields
	if result.DelayMinutes != nil {
		state.DelayMinutes = types.Int64Value(int64(*result.DelayMinutes))
	} else {
		state.DelayMinutes = types.Int64Null()
	}

	if result.RecurringIntervalMinutes != nil {
		state.RecurringIntervalMinutes = types.Int64Value(int64(*result.RecurringIntervalMinutes))
	} else {
		state.RecurringIntervalMinutes = types.Int64Null()
	}

	if result.MaxRecurrences != nil {
		state.MaxRecurrences = types.Int64Value(int64(*result.MaxRecurrences))
	} else {
		state.MaxRecurrences = types.Int64Null()
	}

	state.NotifyOnCall = types.BoolValue(result.NotifyOnCall)
	if result.OnCallPolicyID != nil {
		state.OnCallPolicyID = types.Int64Value(int64(*result.OnCallPolicyID))
	} else {
		state.OnCallPolicyID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alertResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Error(ctx, "alert update is not supported")
	resp.Diagnostics.AddError(
		"Alert Update Not Supported",
		"Alerts cannot be updated directly. They are managed through the statusdrift_monitor resource's alert block.",
	)
}

func (r *alertResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Error(ctx, "alert delete is not supported")
	resp.Diagnostics.AddError(
		"Alert Delete Not Supported",
		"Alerts cannot be deleted directly. They are managed through the statusdrift_monitor resource's alert block.",
	)
}

func (r *alertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Could not parse alert ID %q: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
