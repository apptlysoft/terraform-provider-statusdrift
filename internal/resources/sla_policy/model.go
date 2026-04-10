package sla_policy

import "github.com/hashicorp/terraform-plugin-framework/types"

type SlaPolicyModel struct {
	ID                      types.Int64   `tfsdk:"id"`
	Name                    types.String  `tfsdk:"name"`
	UptimeTarget            types.Float64 `tfsdk:"uptime_target"`
	WindowType              types.String  `tfsdk:"window_type"`
	Scope                   types.String  `tfsdk:"scope"`
	ResponseTimeSlaEnabled  types.Bool    `tfsdk:"response_time_sla_enabled"`
	ResponseTimeThresholdMs types.Int64   `tfsdk:"response_time_threshold_ms"`
	Enabled                 types.Bool    `tfsdk:"enabled"`
	MonitorIDs              types.List    `tfsdk:"monitor_ids"`
	TagIDs                  types.List    `tfsdk:"tag_ids"`
	GroupIDs                types.List    `tfsdk:"group_ids"`
	CreatedAt               types.String  `tfsdk:"created_at"`
	UpdatedAt               types.String  `tfsdk:"updated_at"`
}
