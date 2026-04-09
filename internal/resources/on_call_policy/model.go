package on_call_policy

import "github.com/hashicorp/terraform-plugin-framework/types"

type OnCallPolicyModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	ScopeType      types.String `tfsdk:"scope_type"`
	MonitorGroupID types.Int64  `tfsdk:"monitor_group_id"`
	Steps          types.List   `tfsdk:"step"`
}

type OnCallPolicyStepModel struct {
	Position             types.Int64  `tfsdk:"position"`
	TargetType           types.String `tfsdk:"target_type"`
	ScheduleID           types.Int64  `tfsdk:"schedule_id"`
	OrganizationMemberID types.Int64  `tfsdk:"organization_member_id"`
	DelayMinutes         types.Int64  `tfsdk:"delay_minutes"`
}
