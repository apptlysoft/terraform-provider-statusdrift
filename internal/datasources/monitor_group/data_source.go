package monitor_group

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

var _ datasource.DataSource = &monitorGroupDataSource{}

type monitorGroupDataSource struct {
	client *apiclient.Client
}

type monitorGroupDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewDataSource() datasource.DataSource {
	return &monitorGroupDataSource{}
}

func (d *monitorGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_group"
}

func (d *monitorGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single monitor group by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the monitor group.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the monitor group.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (d *monitorGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *monitorGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state monitorGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var group *apiclient.MonitorGroup

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		groups, err := d.client.ListMonitorGroupsWithSearch(name)
		if err != nil {
			resp.Diagnostics.AddError("Error searching monitor groups", err.Error())
			return
		}
		var matches []apiclient.MonitorGroup
		for _, g := range groups {
			if g.Name == name {
				matches = append(matches, g)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No monitor group found", fmt.Sprintf("No monitor group found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple monitor groups found", fmt.Sprintf("Multiple monitor groups found with name %q, use id instead.", name))
			return
		}
		group = &matches[0]
	} else {
		result, err := d.client.GetMonitorGroup(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading monitor group", err.Error())
			return
		}
		group = result
	}

	state.ID = types.Int64Value(int64(group.ID))
	state.Name = types.StringValue(group.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
