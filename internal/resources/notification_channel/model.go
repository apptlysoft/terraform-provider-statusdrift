package notification_channel

import "github.com/hashicorp/terraform-plugin-framework/types"

type NotificationChannelModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Config  types.Map    `tfsdk:"config"`
	Emails  types.List   `tfsdk:"emails"`
}
