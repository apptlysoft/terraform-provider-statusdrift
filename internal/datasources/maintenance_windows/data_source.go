package maintenance_windows

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &maintenanceWindowsDataSource{}

type maintenanceWindowsDataSource struct {
	client *apiclient.Client
}

type maintenanceWindowsDataSourceModel struct {
	ID                 types.String             `tfsdk:"id"`
	MaintenanceWindows []maintenanceWindowModel `tfsdk:"maintenance_windows"`
}

type maintenanceWindowModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Repeat     types.String `tfsdk:"repeat"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	TargetType types.String `tfsdk:"target_type"`
	Timezone   types.String `tfsdk:"timezone"`
}

func NewDataSource() datasource.DataSource {
	return &maintenanceWindowsDataSource{}
}

func (d *maintenanceWindowsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_windows"
}

func (d *maintenanceWindowsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all maintenance windows.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"maintenance_windows": schema.ListNestedAttribute{
				Description: "List of maintenance windows.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the maintenance window.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the maintenance window.",
							Computed:    true,
						},
						"repeat": schema.StringAttribute{
							Description: "Repeat schedule type.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the maintenance window is enabled.",
							Computed:    true,
						},
						"target_type": schema.StringAttribute{
							Description: "Target type.",
							Computed:    true,
						},
						"timezone": schema.StringAttribute{
							Description: "Timezone.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *maintenanceWindowsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *maintenanceWindowsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	windows, err := d.client.ListMaintenanceWindows()
	if err != nil {
		resp.Diagnostics.AddError("Error listing maintenance windows", err.Error())
		return
	}

	state := maintenanceWindowsDataSourceModel{
		ID:                 types.StringValue("placeholder"),
		MaintenanceWindows: make([]maintenanceWindowModel, len(windows)),
	}

	for i, w := range windows {
		state.MaintenanceWindows[i] = maintenanceWindowModel{
			ID:         types.Int64Value(int64(w.ID)),
			Name:       types.StringValue(w.Name),
			Repeat:     types.StringValue(w.Repeat),
			Enabled:    types.BoolValue(w.Enabled),
			TargetType: types.StringValue(w.TargetType),
			Timezone:   types.StringValue(w.Timezone),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
