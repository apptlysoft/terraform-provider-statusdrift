package status_page_section

import "github.com/hashicorp/terraform-plugin-framework/types"

type StatusPageSectionModel struct {
	ID           types.Int64  `tfsdk:"id"`
	StatusPageID types.Int64  `tfsdk:"status_page_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Position     types.Int64  `tfsdk:"position"`
	Expanded     types.Bool   `tfsdk:"expanded"`
	Enabled      types.Bool   `tfsdk:"enabled"`
}
