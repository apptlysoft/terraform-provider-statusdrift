package on_call_policy

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

var _ datasource.DataSource = &onCallPolicyDataSource{}

type onCallPolicyDataSource struct {
	client *apiclient.Client
}

type onCallPolicyDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	ScopeType      types.String `tfsdk:"scope_type"`
	MonitorGroupID types.Int64  `tfsdk:"monitor_group_id"`
	Steps          types.List   `tfsdk:"steps"`
}

var stepAttrTypes = map[string]attr.Type{
	"position":               types.Int64Type,
	"target_type":            types.StringType,
	"schedule_id":            types.Int64Type,
	"organization_member_id": types.Int64Type,
	"delay_minutes":          types.Int64Type,
}

func NewDataSource() datasource.DataSource {
	return &onCallPolicyDataSource{}
}

func (d *onCallPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_policy"
}

func (d *onCallPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single on-call policy by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the on-call policy.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the on-call policy.",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the policy is active.",
				Computed:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the default organization policy.",
				Computed:    true,
			},
			"scope_type": schema.StringAttribute{
				Description: "Policy scope type.",
				Computed:    true,
			},
			"monitor_group_id": schema.Int64Attribute{
				Description: "Monitor group ID for group-scoped policies.",
				Computed:    true,
			},
			"steps": schema.ListNestedAttribute{
				Description: "Escalation steps.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"position": schema.Int64Attribute{
							Description: "Step position.",
							Computed:    true,
						},
						"target_type": schema.StringAttribute{
							Description: "Target type: schedule or member.",
							Computed:    true,
						},
						"schedule_id": schema.Int64Attribute{
							Description: "On-call schedule ID.",
							Computed:    true,
						},
						"organization_member_id": schema.Int64Attribute{
							Description: "Organization member ID.",
							Computed:    true,
						},
						"delay_minutes": schema.Int64Attribute{
							Description: "Delay in minutes before this step.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *onCallPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *onCallPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state onCallPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy *apiclient.OnCallPolicy

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		policies, err := d.client.ListOnCallPolicies()
		if err != nil {
			resp.Diagnostics.AddError("Error listing on-call policies", err.Error())
			return
		}
		var matches []apiclient.OnCallPolicy
		for _, p := range policies {
			if p.Name == name {
				matches = append(matches, p)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No on-call policy found", fmt.Sprintf("No on-call policy found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple on-call policies found", fmt.Sprintf("Multiple on-call policies found with name %q, use id instead.", name))
			return
		}
		policy = &matches[0]
	} else {
		result, err := d.client.GetOnCallPolicy(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading on-call policy", err.Error())
			return
		}
		policy = result
	}

	state.ID = types.Int64Value(int64(policy.ID))
	state.Name = types.StringValue(policy.Name)
	state.Enabled = types.BoolValue(policy.Enabled)
	state.IsDefault = types.BoolValue(policy.IsDefault)
	state.ScopeType = types.StringValue(policy.ScopeType)

	if policy.MonitorGroup != nil {
		state.MonitorGroupID = types.Int64Value(int64(policy.MonitorGroup.ID))
	} else {
		state.MonitorGroupID = types.Int64Null()
	}

	// Steps
	sorted := make([]apiclient.OnCallPolicyStepRef, len(policy.Steps))
	copy(sorted, policy.Steps)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Position < sorted[j].Position
	})

	if len(sorted) > 0 {
		stepValues := make([]attr.Value, len(sorted))
		for i, s := range sorted {
			attrs := map[string]attr.Value{
				"position":      types.Int64Value(int64(s.Position)),
				"target_type":   types.StringValue(s.TargetType),
				"delay_minutes": types.Int64Value(int64(s.DelayMinutes)),
			}
			if s.Schedule != nil {
				attrs["schedule_id"] = types.Int64Value(int64(s.Schedule.ID))
			} else {
				attrs["schedule_id"] = types.Int64Null()
			}
			if s.OrganizationMember != nil {
				attrs["organization_member_id"] = types.Int64Value(int64(s.OrganizationMember.ID))
			} else {
				attrs["organization_member_id"] = types.Int64Null()
			}
			objVal, diags := types.ObjectValue(stepAttrTypes, attrs)
			resp.Diagnostics.Append(diags...)
			stepValues[i] = objVal
		}
		stepObjType := types.ObjectType{AttrTypes: stepAttrTypes}
		listVal, diags := types.ListValue(stepObjType, stepValues)
		resp.Diagnostics.Append(diags...)
		state.Steps = listVal
	} else {
		stepObjType := types.ObjectType{AttrTypes: stepAttrTypes}
		state.Steps = types.ListNull(stepObjType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
