package on_call_schedule

import "github.com/hashicorp/terraform-plugin-framework/types"

type OnCallScheduleModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ScopeType        types.String `tfsdk:"scope_type"`
	MonitorGroupID   types.Int64  `tfsdk:"monitor_group_id"`
	Timezone         types.String `tfsdk:"timezone"`
	Frequency        types.String `tfsdk:"frequency"`
	RotationInterval types.Int64  `tfsdk:"rotation_interval"`
	StartsAt         types.String `tfsdk:"starts_at"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	ParticipantIDs   types.List   `tfsdk:"participant_ids"`
}
