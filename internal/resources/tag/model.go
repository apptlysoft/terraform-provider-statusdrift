package tag

import "github.com/hashicorp/terraform-plugin-framework/types"

type TagModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Color types.String `tfsdk:"color"`
}
