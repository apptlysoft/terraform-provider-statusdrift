package alert

import "github.com/hashicorp/terraform-plugin-framework/types"

type AlertModel struct {
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
