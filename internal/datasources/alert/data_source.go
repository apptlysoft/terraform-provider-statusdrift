package alert

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &alertDataSource{}

type alertDataSource struct {
	client *apiclient.Client
}

type alertDataSourceModel struct {
	ID                       types.Int64 `tfsdk:"id"`
	MonitorID                types.Int64 `tfsdk:"monitor_id"`
	NotificationChannelIDs   types.List  `tfsdk:"notification_channel_ids"`
	Conditions               types.List  `tfsdk:"conditions"`
	DelayMinutes             types.Int64 `tfsdk:"delay_minutes"`
	Enabled                  types.Bool  `tfsdk:"enabled"`
	RecurringIntervalMinutes types.Int64 `tfsdk:"recurring_interval_minutes"`
	MaxRecurrences           types.Int64 `tfsdk:"max_recurrences"`
	NotifyOnCall             types.Bool  `tfsdk:"notify_on_call"`
	OnCallPolicyID           types.Int64 `tfsdk:"on_call_policy_id"`
}

func NewDataSource() datasource.DataSource {
	return &alertDataSource{}
}

func (d *alertDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (d *alertDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single alert by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the alert.",
				Required:    true,
			},
			"monitor_id": schema.Int64Attribute{
				Description: "The ID of the monitor this alert belongs to.",
				Computed:    true,
			},
			"notification_channel_ids": schema.ListAttribute{
				Description: "List of notification channel IDs.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"conditions": schema.ListAttribute{
				Description: "List of alert conditions.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"delay_minutes": schema.Int64Attribute{
				Description: "Delay in minutes before triggering the alert.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert is enabled.",
				Computed:    true,
			},
			"recurring_interval_minutes": schema.Int64Attribute{
				Description: "Recurring interval in minutes.",
				Computed:    true,
			},
			"max_recurrences": schema.Int64Attribute{
				Description: "Maximum number of recurrences.",
				Computed:    true,
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

func (d *alertDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *alertDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state alertDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := d.client.GetAlert(int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert", err.Error())
		return
	}

	state.ID = types.Int64Value(int64(alert.ID))
	state.MonitorID = types.Int64Value(int64(alert.MonitorID))
	state.Enabled = types.BoolValue(alert.Enabled)

	channelList, diags := types.ListValueFrom(ctx, types.Int64Type, alert.NotificationChannelIDs)
	resp.Diagnostics.Append(diags...)
	state.NotificationChannelIDs = channelList

	condList, diags := types.ListValueFrom(ctx, types.StringType, alert.Conditions)
	resp.Diagnostics.Append(diags...)
	state.Conditions = condList

	if alert.DelayMinutes != nil {
		state.DelayMinutes = types.Int64Value(int64(*alert.DelayMinutes))
	} else {
		state.DelayMinutes = types.Int64Null()
	}
	if alert.RecurringIntervalMinutes != nil {
		state.RecurringIntervalMinutes = types.Int64Value(int64(*alert.RecurringIntervalMinutes))
	} else {
		state.RecurringIntervalMinutes = types.Int64Null()
	}
	if alert.MaxRecurrences != nil {
		state.MaxRecurrences = types.Int64Value(int64(*alert.MaxRecurrences))
	} else {
		state.MaxRecurrences = types.Int64Null()
	}

	state.NotifyOnCall = types.BoolValue(alert.NotifyOnCall)
	if alert.OnCallPolicyID != nil {
		state.OnCallPolicyID = types.Int64Value(int64(*alert.OnCallPolicyID))
	} else {
		state.OnCallPolicyID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
