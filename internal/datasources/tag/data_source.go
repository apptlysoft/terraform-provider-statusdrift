package tag

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &tagDataSource{}

type tagDataSource struct {
	client *apiclient.Client
}

type tagDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Color types.String `tfsdk:"color"`
}

func NewDataSource() datasource.DataSource {
	return &tagDataSource{}
}

func (d *tagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (d *tagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single tag by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the tag.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the tag.",
				Optional:    true,
				Computed:    true,
			},
			"color": schema.StringAttribute{
				Description: "The color of the tag (hex code).",
				Computed:    true,
			},
		},
	}
}

func (d *tagDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *tagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tagDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tag *apiclient.Tag

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		tags, err := d.client.ListTagsWithSearch(name)
		if err != nil {
			resp.Diagnostics.AddError("Error searching tags", err.Error())
			return
		}
		var matches []apiclient.Tag
		for _, t := range tags {
			if t.Name == name {
				matches = append(matches, t)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No tag found", fmt.Sprintf("No tag found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple tags found", fmt.Sprintf("Multiple tags found with name %q, use id instead.", name))
			return
		}
		tag = &matches[0]
	} else {
		result, err := d.client.GetTag(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading tag", err.Error())
			return
		}
		tag = result
	}

	state.ID = types.Int64Value(int64(tag.ID))
	state.Name = types.StringValue(tag.Name)
	if tag.Color != nil {
		state.Color = types.StringValue(*tag.Color)
	} else {
		state.Color = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
