package monitor_group

import "github.com/hashicorp/terraform-plugin-framework/types"

type MonitorGroupModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
