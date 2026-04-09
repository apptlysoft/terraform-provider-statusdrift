package incident

import "github.com/hashicorp/terraform-plugin-framework/types"

type IncidentModel struct {
	ID        types.Int64  `tfsdk:"id"`
	MonitorID types.Int64  `tfsdk:"monitor_id"`
	Message   types.String `tfsdk:"message"`
	Priority  types.String `tfsdk:"priority"`
	Resolved  types.Bool   `tfsdk:"resolved"`
}
