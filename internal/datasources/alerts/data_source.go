package alerts

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &alertsDataSource{}

type alertsDataSource struct {
	client *apiclient.Client
}

type alertsDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Alerts []alertModel `tfsdk:"alerts"`
}

type alertModel struct {
	ID             types.Int64 `tfsdk:"id"`
	MonitorID      types.Int64 `tfsdk:"monitor_id"`
	Enabled        types.Bool  `tfsdk:"enabled"`
	Conditions     types.List  `tfsdk:"conditions"`
	NotifyOnCall   types.Bool  `tfsdk:"notify_on_call"`
	OnCallPolicyID types.Int64 `tfsdk:"on_call_policy_id"`
}

func NewDataSource() datasource.DataSource {
	return &alertsDataSource{}
}

func (d *alertsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerts"
}

func (d *alertsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all alerts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"alerts": schema.ListNestedAttribute{
				Description: "List of alerts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the alert.",
							Computed:    true,
						},
						"monitor_id": schema.Int64Attribute{
							Description: "The ID of the monitor this alert belongs to.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the alert is enabled.",
							Computed:    true,
						},
						"conditions": schema.ListAttribute{
							Description: "List of alert conditions.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"notify_on_call": schema.BoolAttribute{
							Description: "Whether this alert triggers the on-call escalation policy.",
							Computed:    true,
						},
						"on_call_policy_id": schema.Int64Attribute{
							Description: "On-call policy ID used for this alert.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *alertsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alertsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	alerts, err := d.client.ListAlerts()
	if err != nil {
		resp.Diagnostics.AddError("Error listing alerts", err.Error())
		return
	}

	state := alertsDataSourceModel{
		ID:     types.StringValue("placeholder"),
		Alerts: make([]alertModel, len(alerts)),
	}

	for i, a := range alerts {
		condList, diags := types.ListValueFrom(ctx, types.StringType, a.Conditions)
		resp.Diagnostics.Append(diags...)
		var onCallPolicyID types.Int64
		if a.OnCallPolicyID != nil {
			onCallPolicyID = types.Int64Value(int64(*a.OnCallPolicyID))
		} else {
			onCallPolicyID = types.Int64Null()
		}
		state.Alerts[i] = alertModel{
			ID:             types.Int64Value(int64(a.ID)),
			MonitorID:      types.Int64Value(int64(a.MonitorID)),
			Enabled:        types.BoolValue(a.Enabled),
			Conditions:     condList,
			NotifyOnCall:   types.BoolValue(a.NotifyOnCall),
			OnCallPolicyID: onCallPolicyID,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
