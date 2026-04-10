package on_call_schedule

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &onCallScheduleDataSource{}

type onCallScheduleDataSource struct {
	client *apiclient.Client
}

type onCallScheduleDataSourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ScopeType        types.String `tfsdk:"scope_type"`
	MonitorGroupID   types.Int64  `tfsdk:"monitor_group_id"`
	Timezone         types.String `tfsdk:"timezone"`
	Frequency        types.String `tfsdk:"frequency"`
	RotationInterval types.Int64  `tfsdk:"rotation_interval"`
	StartsAt         types.String `tfsdk:"starts_at"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	ParticipantIDs   types.List   `tfsdk:"participant_ids"`
}

func NewDataSource() datasource.DataSource {
	return &onCallScheduleDataSource{}
}

func (d *onCallScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_schedule"
}

func (d *onCallScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single on-call schedule by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the on-call schedule.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the on-call schedule.",
				Optional:    true,
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
			"rotation_interval": schema.Int64Attribute{
				Description: "Rotation interval.",
				Computed:    true,
			},
			"starts_at": schema.StringAttribute{
				Description: "Schedule start time.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the schedule is active.",
				Computed:    true,
			},
			"participant_ids": schema.ListAttribute{
				Description: "Ordered list of organization member IDs in the rotation.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
		},
	}
}

func (d *onCallScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *onCallScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state onCallScheduleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var schedule *apiclient.OnCallSchedule

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		schedules, err := d.client.ListOnCallSchedules()
		if err != nil {
			resp.Diagnostics.AddError("Error listing on-call schedules", err.Error())
			return
		}
		var matches []apiclient.OnCallSchedule
		for _, s := range schedules {
			if s.Name == name {
				matches = append(matches, s)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No on-call schedule found", fmt.Sprintf("No on-call schedule found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple on-call schedules found", fmt.Sprintf("Multiple on-call schedules found with name %q, use id instead.", name))
			return
		}
		schedule = &matches[0]
	} else {
		result, err := d.client.GetOnCallSchedule(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading on-call schedule", err.Error())
			return
		}
		schedule = result
	}

	state.ID = types.Int64Value(int64(schedule.ID))
	state.Name = types.StringValue(schedule.Name)
	state.ScopeType = types.StringValue(schedule.ScopeType)
	if schedule.MonitorGroup != nil {
		state.MonitorGroupID = types.Int64Value(int64(schedule.MonitorGroup.ID))
	} else {
		state.MonitorGroupID = types.Int64Null()
	}
	state.Timezone = types.StringValue(schedule.Timezone)
	state.Frequency = types.StringValue(schedule.Frequency)
	state.RotationInterval = types.Int64Value(int64(schedule.RotationInterval))
	state.StartsAt = types.StringValue(schedule.StartsAt)
	state.Enabled = types.BoolValue(schedule.Enabled)

	sorted := make([]apiclient.OnCallParticipant, len(schedule.Participants))
	copy(sorted, schedule.Participants)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Position < sorted[j].Position
	})

	if len(sorted) > 0 {
		ids := make([]attr.Value, len(sorted))
		for i, p := range sorted {
			ids[i] = types.Int64Value(int64(p.OrganizationMember.ID))
		}
		listVal, diags := types.ListValue(types.Int64Type, ids)
		resp.Diagnostics.Append(diags...)
		state.ParticipantIDs = listVal
	} else {
		state.ParticipantIDs = types.ListNull(types.Int64Type)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
