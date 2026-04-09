package on_call_schedule_override

import "github.com/hashicorp/terraform-plugin-framework/types"

type OnCallScheduleOverrideModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	ScheduleID          types.Int64  `tfsdk:"schedule_id"`
	ReplacementMemberID types.Int64  `tfsdk:"replacement_member_id"`
	StartsAt            types.String `tfsdk:"starts_at"`
	EndsAt              types.String `tfsdk:"ends_at"`
	Note                types.String `tfsdk:"note"`
}
