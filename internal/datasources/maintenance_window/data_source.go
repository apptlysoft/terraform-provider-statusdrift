package maintenance_window

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

var _ datasource.DataSource = &maintenanceWindowDataSource{}

type maintenanceWindowDataSource struct {
	client *apiclient.Client
}

type maintenanceWindowDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Repeat         types.String `tfsdk:"repeat"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	TargetType     types.String `tfsdk:"target_type"`
	TargetID       types.Int64  `tfsdk:"target_id"`
	StartDateTime  types.String `tfsdk:"start_date_time"`
	EndDateTime    types.String `tfsdk:"end_date_time"`
	StartTime      types.String `tfsdk:"start_time"`
	EndTime        types.String `tfsdk:"end_time"`
	Timezone       types.String `tfsdk:"timezone"`
	WeeklyDays     types.List   `tfsdk:"weekly_days"`
	MonthlyDayType types.String `tfsdk:"monthly_day_type"`
	MonthlyDay     types.Int64  `tfsdk:"monthly_day"`
}

func NewDataSource() datasource.DataSource {
	return &maintenanceWindowDataSource{}
}

func (d *maintenanceWindowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_window"
}

func (d *maintenanceWindowDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single maintenance window by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the maintenance window.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the maintenance window.",
				Optional:    true,
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
				Description: "Target type (monitor, monitor_group, tag).",
				Computed:    true,
			},
			"target_id": schema.Int64Attribute{
				Description: "The ID of the target resource.",
				Computed:    true,
			},
			"start_date_time": schema.StringAttribute{
				Description: "Start date/time for one-time maintenance.",
				Computed:    true,
			},
			"end_date_time": schema.StringAttribute{
				Description: "End date/time for one-time maintenance.",
				Computed:    true,
			},
			"start_time": schema.StringAttribute{
				Description: "Start time for recurring maintenance.",
				Computed:    true,
			},
			"end_time": schema.StringAttribute{
				Description: "End time for recurring maintenance.",
				Computed:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone for the maintenance window.",
				Computed:    true,
			},
			"weekly_days": schema.ListAttribute{
				Description: "Days of the week for weekly maintenance.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"monthly_day_type": schema.StringAttribute{
				Description: "Monthly day type.",
				Computed:    true,
			},
			"monthly_day": schema.Int64Attribute{
				Description: "Day of the month for monthly maintenance.",
				Computed:    true,
			},
		},
	}
}

func (d *maintenanceWindowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *maintenanceWindowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state maintenanceWindowDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mw *apiclient.MaintenanceWindow

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		windows, err := d.client.ListMaintenanceWindows()
		if err != nil {
			resp.Diagnostics.AddError("Error listing maintenance windows", err.Error())
			return
		}
		var matches []apiclient.MaintenanceWindow
		for _, w := range windows {
			if w.Name == name {
				matches = append(matches, w)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No maintenance window found", fmt.Sprintf("No maintenance window found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple maintenance windows found", fmt.Sprintf("Multiple maintenance windows found with name %q, use id instead.", name))
			return
		}
		mw = &matches[0]
	} else {
		result, err := d.client.GetMaintenanceWindow(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading maintenance window", err.Error())
			return
		}
		mw = result
	}

	state.ID = types.Int64Value(int64(mw.ID))
	state.Name = types.StringValue(mw.Name)
	state.Repeat = types.StringValue(mw.Repeat)
	state.Enabled = types.BoolValue(mw.Enabled)
	state.TargetType = types.StringValue(mw.TargetType)
	state.Timezone = types.StringValue(mw.Timezone)

	weeklyDaysList, diags := types.ListValueFrom(ctx, types.StringType, mw.WeeklyDays)
	resp.Diagnostics.Append(diags...)
	state.WeeklyDays = weeklyDaysList

	if mw.TargetID != nil {
		state.TargetID = types.Int64Value(int64(*mw.TargetID))
	} else {
		state.TargetID = types.Int64Null()
	}
	if mw.StartDateTime != nil {
		state.StartDateTime = types.StringValue(*mw.StartDateTime)
	} else {
		state.StartDateTime = types.StringNull()
	}
	if mw.EndDateTime != nil {
		state.EndDateTime = types.StringValue(*mw.EndDateTime)
	} else {
		state.EndDateTime = types.StringNull()
	}
	if mw.StartTime != nil {
		state.StartTime = types.StringValue(*mw.StartTime)
	} else {
		state.StartTime = types.StringNull()
	}
	if mw.EndTime != nil {
		state.EndTime = types.StringValue(*mw.EndTime)
	} else {
		state.EndTime = types.StringNull()
	}
	if mw.MonthlyDayType != nil {
		state.MonthlyDayType = types.StringValue(*mw.MonthlyDayType)
	} else {
		state.MonthlyDayType = types.StringNull()
	}
	if mw.MonthlyDay != nil {
		state.MonthlyDay = types.Int64Value(int64(*mw.MonthlyDay))
	} else {
		state.MonthlyDay = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
