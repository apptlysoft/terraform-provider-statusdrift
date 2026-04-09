package status_page_announcement

import "github.com/hashicorp/terraform-plugin-framework/types"

type StatusPageAnnouncementModel struct {
	ID           types.Int64  `tfsdk:"id"`
	StatusPageID types.Int64  `tfsdk:"status_page_id"`
	Title        types.String `tfsdk:"title"`
	Content      types.String `tfsdk:"content"`
	Type         types.String `tfsdk:"type"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	StartsAt     types.String `tfsdk:"starts_at"`
	EndsAt       types.String `tfsdk:"ends_at"`
}
