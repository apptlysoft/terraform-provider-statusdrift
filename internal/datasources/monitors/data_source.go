package monitors

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &monitorsDataSource{}

type monitorsDataSource struct {
	client *apiclient.Client
}

type monitorsDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Monitors []monitorModel `tfsdk:"monitors"`
}

type monitorModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Interval types.Int64  `tfsdk:"interval"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Status   types.String `tfsdk:"status"`
}

func NewDataSource() datasource.DataSource {
	return &monitorsDataSource{}
}

func (d *monitorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

func (d *monitorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all monitors.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"monitors": schema.ListNestedAttribute{
				Description: "List of monitors.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the monitor.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the monitor.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the monitor.",
							Computed:    true,
						},
						"interval": schema.Int64Attribute{
							Description: "Check interval in seconds.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the monitor is enabled.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the monitor.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *monitorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *monitorsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	monitors, err := d.client.ListMonitors()
	if err != nil {
		resp.Diagnostics.AddError("Error listing monitors", err.Error())
		return
	}

	state := monitorsDataSourceModel{
		ID:       types.StringValue("placeholder"),
		Monitors: make([]monitorModel, len(monitors)),
	}

	for i, m := range monitors {
		state.Monitors[i] = monitorModel{
			ID:       types.Int64Value(int64(m.ID)),
			Name:     types.StringValue(m.Name),
			Type:     types.StringValue(m.Type),
			Interval: types.Int64Value(int64(m.Interval)),
			Enabled:  types.BoolValue(m.Enabled),
		}
		if m.Status != nil {
			state.Monitors[i].Status = types.StringValue(*m.Status)
		} else {
			state.Monitors[i].Status = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
