package on_call_schedules

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &onCallSchedulesDataSource{}

type onCallSchedulesDataSource struct {
	client *apiclient.Client
}

type onCallSchedulesDataSourceModel struct {
	ID        types.String          `tfsdk:"id"`
	Schedules []onCallScheduleModel `tfsdk:"on_call_schedules"`
}

type onCallScheduleModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ScopeType      types.String `tfsdk:"scope_type"`
	MonitorGroupID types.Int64  `tfsdk:"monitor_group_id"`
	Timezone       types.String `tfsdk:"timezone"`
	Frequency      types.String `tfsdk:"frequency"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}

func NewDataSource() datasource.DataSource {
	return &onCallSchedulesDataSource{}
}

func (d *onCallSchedulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_schedules"
}

func (d *onCallSchedulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all on-call schedules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"on_call_schedules": schema.ListNestedAttribute{
				Description: "List of on-call schedules.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the on-call schedule.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the on-call schedule.",
							Computed:    true,
						},
						"scope_type": schema.StringAttribute{
							Description: "Schedule scope: org_default or monitor_group.",
							Computed:    true,
						},
						"monitor_group_id": schema.Int64Attribute{
							Description: "The monitor group ID this schedule is scoped to, if any.",
							Computed:    true,
						},
						"timezone": schema.StringAttribute{
							Description: "IANA timezone.",
							Computed:    true,
						},
						"frequency": schema.StringAttribute{
							Description: "Rotation frequency.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the schedule is active.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *onCallSchedulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *apiclient.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *onCallSchedulesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	schedules, err := d.client.ListOnCallSchedules()
	if err != nil {
		resp.Diagnostics.AddError("Error listing on-call schedules", err.Error())
		return
	}

	state := onCallSchedulesDataSourceModel{
		ID:        types.StringValue("placeholder"),
		Schedules: make([]onCallScheduleModel, len(schedules)),
	}

	for i, s := range schedules {
		var monitorGroupID types.Int64
		if s.MonitorGroup != nil {
			monitorGroupID = types.Int64Value(int64(s.MonitorGroup.ID))
		} else {
			monitorGroupID = types.Int64Null()
		}
		state.Schedules[i] = onCallScheduleModel{
			ID:             types.Int64Value(int64(s.ID)),
			Name:           types.StringValue(s.Name),
			ScopeType:      types.StringValue(s.ScopeType),
			MonitorGroupID: monitorGroupID,
			Timezone:       types.StringValue(s.Timezone),
			Frequency:      types.StringValue(s.Frequency),
			Enabled:        types.BoolValue(s.Enabled),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
