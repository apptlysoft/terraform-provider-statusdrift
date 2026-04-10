package sla_policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &slaPolicyDataSource{}

type slaPolicyDataSource struct {
	client *apiclient.Client
}

type slaPolicyDataSourceModel struct {
	ID                      types.Int64   `tfsdk:"id"`
	Name                    types.String  `tfsdk:"name"`
	UptimeTarget            types.Float64 `tfsdk:"uptime_target"`
	WindowType              types.String  `tfsdk:"window_type"`
	Scope                   types.String  `tfsdk:"scope"`
	ResponseTimeSlaEnabled  types.Bool    `tfsdk:"response_time_sla_enabled"`
	ResponseTimeThresholdMs types.Int64   `tfsdk:"response_time_threshold_ms"`
	Enabled                 types.Bool    `tfsdk:"enabled"`
	MonitorIDs              types.List    `tfsdk:"monitor_ids"`
	TagIDs                  types.List    `tfsdk:"tag_ids"`
	GroupIDs                types.List    `tfsdk:"group_ids"`
	CreatedAt               types.String  `tfsdk:"created_at"`
	UpdatedAt               types.String  `tfsdk:"updated_at"`
}

func NewDataSource() datasource.DataSource {
	return &slaPolicyDataSource{}
}

func (d *slaPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sla_policy"
}

func (d *slaPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single SLA policy by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the SLA policy.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the SLA policy.",
				Optional:    true,
				Computed:    true,
			},
			"uptime_target": schema.Float64Attribute{
				Description: "Target uptime percentage.",
				Computed:    true,
			},
			"window_type": schema.StringAttribute{
				Description: "SLA evaluation window type.",
				Computed:    true,
			},
			"scope": schema.StringAttribute{
				Description: "Policy scope: org or group.",
				Computed:    true,
			},
			"response_time_sla_enabled": schema.BoolAttribute{
				Description: "Whether response time SLA tracking is enabled.",
				Computed:    true,
			},
			"response_time_threshold_ms": schema.Int64Attribute{
				Description: "Response time threshold in milliseconds.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the SLA policy is active.",
				Computed:    true,
			},
			"monitor_ids": schema.ListAttribute{
				Description: "List of monitor IDs directly covered by this policy.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"tag_ids": schema.ListAttribute{
				Description: "List of tag IDs for dynamic monitor selection.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"group_ids": schema.ListAttribute{
				Description: "List of monitor group IDs covered by this policy.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"created_at": schema.StringAttribute{
				Description: "When the SLA policy was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the SLA policy was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *slaPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *slaPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state slaPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy *apiclient.SlaPolicy

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		policies, err := d.client.ListSlaPoliciesWithSearch(name)
		if err != nil {
			resp.Diagnostics.AddError("Error searching SLA policies", err.Error())
			return
		}
		var matches []apiclient.SlaPolicy
		for _, p := range policies {
			if p.Name == name {
				matches = append(matches, p)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No SLA policy found", fmt.Sprintf("No SLA policy found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple SLA policies found", fmt.Sprintf("Multiple SLA policies found with name %q, use id instead.", name))
			return
		}
		policy = &matches[0]
	} else {
		result, err := d.client.GetSlaPolicy(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading SLA policy", err.Error())
			return
		}
		policy = result
	}

	state.ID = types.Int64Value(int64(policy.ID))
	state.Name = types.StringValue(policy.Name)
	state.UptimeTarget = types.Float64Value(policy.UptimeTarget)
	state.WindowType = types.StringValue(policy.WindowType)
	state.Scope = types.StringValue(policy.Scope)
	state.ResponseTimeSlaEnabled = types.BoolValue(policy.ResponseTimeSlaEnabled)
	state.Enabled = types.BoolValue(policy.Enabled)
	state.CreatedAt = types.StringValue(policy.CreatedAt)
	state.UpdatedAt = types.StringValue(policy.UpdatedAt)

	if policy.ResponseTimeThresholdMs != nil {
		state.ResponseTimeThresholdMs = types.Int64Value(int64(*policy.ResponseTimeThresholdMs))
	} else {
		state.ResponseTimeThresholdMs = types.Int64Null()
	}

	state.MonitorIDs = refsToIDList(policy.Monitors)
	state.TagIDs = refsToIDList(policy.Tags)
	state.GroupIDs = refsToIDList(policy.Groups)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type idGetter interface {
	GetID() int
}

func refsToIDList[T idGetter](refs []T) types.List {
	if len(refs) == 0 {
		return types.ListNull(types.Int64Type)
	}
	ids := make([]attr.Value, len(refs))
	for i, r := range refs {
		ids[i] = types.Int64Value(int64(r.GetID()))
	}
	listVal, _ := types.ListValue(types.Int64Type, ids)
	return listVal
}
