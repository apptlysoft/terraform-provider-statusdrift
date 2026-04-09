package incidents

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &incidentsDataSource{}

type incidentsDataSource struct {
	client *apiclient.Client
}

type incidentsDataSourceModel struct {
	ID        types.String    `tfsdk:"id"`
	Status    types.String    `tfsdk:"status"`
	Type      types.String    `tfsdk:"type"`
	Priority  types.String    `tfsdk:"priority"`
	MonitorID types.Int64     `tfsdk:"monitor_id"`
	Incidents []incidentModel `tfsdk:"incidents"`
}

type incidentModel struct {
	ID        types.Int64  `tfsdk:"id"`
	MonitorID types.Int64  `tfsdk:"monitor_id"`
	Priority  types.String `tfsdk:"priority"`
	Resolved  types.Bool   `tfsdk:"resolved"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewDataSource() datasource.DataSource {
	return &incidentsDataSource{}
}

func (d *incidentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incidents"
}

func (d *incidentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches incidents, optionally filtered by status, type, priority, or monitor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Filter by incident status: open or resolved.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("open", "resolved"),
				},
			},
			"type": schema.StringAttribute{
				Description: "Filter by incident type: manual or system.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("manual", "system"),
				},
			},
			"priority": schema.StringAttribute{
				Description: "Filter by incident priority: low, medium, high, or critical.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("low", "medium", "high", "critical"),
				},
			},
			"monitor_id": schema.Int64Attribute{
				Description: "Filter by monitor ID.",
				Optional:    true,
			},
			"incidents": schema.ListNestedAttribute{
				Description: "List of incidents.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the incident.",
							Computed:    true,
						},
						"monitor_id": schema.Int64Attribute{
							Description: "The ID of the monitor this incident belongs to.",
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
						"created_at": schema.StringAttribute{
							Description: "When the incident was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *incidentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *incidentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config incidentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters := apiclient.IncidentListFilters{}
	if !config.Status.IsNull() {
		v := config.Status.ValueString()
		filters.Status = &v
	}
	if !config.Type.IsNull() {
		v := config.Type.ValueString()
		filters.Type = &v
	}
	if !config.Priority.IsNull() {
		v := config.Priority.ValueString()
		filters.Priority = &v
	}
	if !config.MonitorID.IsNull() {
		v := int(config.MonitorID.ValueInt64())
		filters.MonitorID = &v
	}

	incidents, err := d.client.ListIncidentsFiltered(filters)
	if err != nil {
		resp.Diagnostics.AddError("Error listing incidents", err.Error())
		return
	}

	state := incidentsDataSourceModel{
		ID:        types.StringValue("placeholder"),
		Status:    config.Status,
		Type:      config.Type,
		Priority:  config.Priority,
		MonitorID: config.MonitorID,
		Incidents: make([]incidentModel, len(incidents)),
	}

	for i, inc := range incidents {
		state.Incidents[i] = incidentModel{
			ID:        types.Int64Value(int64(inc.ID)),
			MonitorID: types.Int64Value(int64(inc.MonitorID)),
			Resolved:  types.BoolValue(inc.Resolved),
			CreatedAt: types.StringValue(inc.CreatedAt),
		}
		if inc.Priority != nil {
			state.Incidents[i].Priority = types.StringValue(*inc.Priority)
		} else {
			state.Incidents[i].Priority = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
