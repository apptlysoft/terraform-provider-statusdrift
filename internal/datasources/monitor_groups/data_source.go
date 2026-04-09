package monitor_groups

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &monitorGroupsDataSource{}

type monitorGroupsDataSource struct {
	client *apiclient.Client
}

type monitorGroupsDataSourceModel struct {
	ID            types.String        `tfsdk:"id"`
	MonitorGroups []monitorGroupModel `tfsdk:"monitor_groups"`
}

type monitorGroupModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewDataSource() datasource.DataSource {
	return &monitorGroupsDataSource{}
}

func (d *monitorGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_groups"
}

func (d *monitorGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all monitor groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"monitor_groups": schema.ListNestedAttribute{
				Description: "List of monitor groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the monitor group.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the monitor group.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *monitorGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *monitorGroupsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	groups, err := d.client.ListMonitorGroups()
	if err != nil {
		resp.Diagnostics.AddError("Error listing monitor groups", err.Error())
		return
	}

	state := monitorGroupsDataSourceModel{
		ID:            types.StringValue("placeholder"),
		MonitorGroups: make([]monitorGroupModel, len(groups)),
	}

	for i, g := range groups {
		state.MonitorGroups[i] = monitorGroupModel{
			ID:   types.Int64Value(int64(g.ID)),
			Name: types.StringValue(g.Name),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
