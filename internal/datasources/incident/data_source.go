package incident

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &incidentDataSource{}

type incidentDataSource struct {
	client *apiclient.Client
}

type incidentDataSourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	MonitorID  types.Int64  `tfsdk:"monitor_id"`
	Message    types.String `tfsdk:"message"`
	Priority   types.String `tfsdk:"priority"`
	Resolved   types.Bool   `tfsdk:"resolved"`
	ResolvedAt types.String `tfsdk:"resolved_at"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func NewDataSource() datasource.DataSource {
	return &incidentDataSource{}
}

func (d *incidentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incident"
}

func (d *incidentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single incident by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the incident.",
				Required:    true,
			},
			"monitor_id": schema.Int64Attribute{
				Description: "The ID of the monitor this incident belongs to.",
				Computed:    true,
			},
			"message": schema.StringAttribute{
				Description: "The incident message.",
				Computed:    true,
			},
			"priority": schema.StringAttribute{
				Description: "The priority of the incident.",
				Computed:    true,
			},
			"resolved": schema.BoolAttribute{
				Description: "Whether the incident is resolved.",
				Computed:    true,
			},
			"resolved_at": schema.StringAttribute{
				Description: "When the incident was resolved.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the incident was created.",
				Computed:    true,
			},
		},
	}
}

func (d *incidentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *incidentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state incidentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	incident, err := d.client.GetIncident(int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading incident", err.Error())
		return
	}

	state.ID = types.Int64Value(int64(incident.ID))
	state.MonitorID = types.Int64Value(int64(incident.MonitorID))
	state.Resolved = types.BoolValue(incident.Resolved)
	state.CreatedAt = types.StringValue(incident.CreatedAt)

	if incident.Message != nil {
		state.Message = types.StringValue(*incident.Message)
	} else {
		state.Message = types.StringNull()
	}
	if incident.Priority != nil {
		state.Priority = types.StringValue(*incident.Priority)
	} else {
		state.Priority = types.StringNull()
	}
	if incident.ResolvedAt != nil {
		state.ResolvedAt = types.StringValue(*incident.ResolvedAt)
	} else {
		state.ResolvedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
