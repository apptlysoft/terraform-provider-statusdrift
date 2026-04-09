package maintenance_window

import "github.com/hashicorp/terraform-plugin-framework/types"

type MaintenanceWindowModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Repeat         types.String `tfsdk:"repeat"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	TargetType     types.String `tfsdk:"target_type"`
	TargetID       types.Int64  `tfsdk:"target_id"`
	StartDateTime  types.String `tfsdk:"start_date_time"`
	EndDateTime    types.String `tfsdk:"end_date_time"`
	StartTime      types.String `tfsdk:"start_time"`
	EndTime        types.String `tfsdk:"end_time"`
	Timezone       types.String `tfsdk:"timezone"`
	WeeklyDays     types.List   `tfsdk:"weekly_days"`
	MonthlyDayType types.String `tfsdk:"monthly_day_type"`
	MonthlyDay     types.Int64  `tfsdk:"monthly_day"`
}
