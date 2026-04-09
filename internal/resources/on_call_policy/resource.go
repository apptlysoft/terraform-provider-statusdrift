package on_call_policy

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                   = &onCallPolicyResource{}
	_ resource.ResourceWithImportState    = &onCallPolicyResource{}
	_ resource.ResourceWithValidateConfig = &onCallPolicyResource{}
)

type onCallPolicyResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &onCallPolicyResource{}
}

func (r *onCallPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_policy"
}

var stepAttrTypes = map[string]attr.Type{
	"position":               types.Int64Type,
	"target_type":            types.StringType,
	"schedule_id":            types.Int64Type,
	"organization_member_id": types.Int64Type,
	"delay_minutes":          types.Int64Type,
}

func (r *onCallPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift on-call escalation policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the on-call policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the on-call policy (max 255 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the on-call policy is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the default organization on-call policy. Only one policy can be default. Only applies to org_default scope.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"scope_type": schema.StringAttribute{
				Description: "Policy scope: org_default (organization-wide) or monitor_group (group-specific).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("org_default"),
				Validators: []validator.String{
					stringvalidator.OneOf("org_default", "monitor_group"),
				},
			},
			"monitor_group_id": schema.Int64Attribute{
				Description: "Monitor group ID. Required when scope_type is monitor_group.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"step": schema.ListNestedBlock{
				Description: "Escalation steps, ordered by position. Steps are executed in order.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"position": schema.Int64Attribute{
							Description: "Step position (1-indexed), computed from block order.",
							Computed:    true,
						},
						"target_type": schema.StringAttribute{
							Description: "Target type: schedule (notify current on-call from schedule) or member (notify specific member).",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("schedule", "member"),
							},
						},
						"schedule_id": schema.Int64Attribute{
							Description: "On-call schedule ID. Required when target_type is schedule.",
							Optional:    true,
						},
						"organization_member_id": schema.Int64Attribute{
							Description: "Organization member ID. Required when target_type is member.",
							Optional:    true,
						},
						"delay_minutes": schema.Int64Attribute{
							Description: "Minutes to wait before escalating to this step (0 = immediate).",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
					},
				},
			},
		},
	}
}

func (r *onCallPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data OnCallPolicyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate monitor_group_id is set when scope_type is monitor_group.
	// Skip when monitor_group_id is unknown (will be resolved at apply time).
	if !data.ScopeType.IsNull() && !data.ScopeType.IsUnknown() && data.ScopeType.ValueString() == "monitor_group" {
		if data.MonitorGroupID.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("monitor_group_id"),
				"Missing Required Attribute",
				"monitor_group_id is required when scope_type is 'monitor_group'.",
			)
		}
	}

	// Validate step targets.
	// Skip unknown values — they reference other resources not yet created.
	if !data.Steps.IsNull() && !data.Steps.IsUnknown() {
		var steps []OnCallPolicyStepModel
		resp.Diagnostics.Append(data.Steps.ElementsAs(ctx, &steps, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for i, step := range steps {
			if step.TargetType.IsNull() || step.TargetType.IsUnknown() {
				continue
			}
			tt := step.TargetType.ValueString()
			if tt == "schedule" && step.ScheduleID.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("step").AtListIndex(i).AtName("schedule_id"),
					"Missing Required Attribute",
					"schedule_id is required when target_type is 'schedule'.",
				)
			}
			if tt == "member" && step.OrganizationMemberID.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("step").AtListIndex(i).AtName("organization_member_id"),
					"Missing Required Attribute",
					"organization_member_id is required when target_type is 'member'.",
				)
			}
		}
	}
}

func (r *onCallPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *onCallPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OnCallPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := modelToInput(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateOnCallPolicy(input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating on-call policy", fmt.Sprintf("Could not create on-call policy: %s", err))
		return
	}

	// If is_default is requested, set it after creation.
	if plan.IsDefault.ValueBool() && !result.IsDefault {
		result, err = r.client.SetDefaultOnCallPolicy(result.ID)
		if err != nil {
			resp.Diagnostics.AddError("Error setting default on-call policy", fmt.Sprintf("Policy created but could not set as default: %s", err))
			return
		}
	}

	apiToModel(result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created on-call policy", map[string]any{"id": result.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *onCallPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OnCallPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	result, err := r.client.GetOnCallPolicy(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "on-call policy not found, removing from state", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading on-call policy", fmt.Sprintf("Could not read on-call policy %d: %s", id, err))
		return
	}

	apiToModel(result, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *onCallPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OnCallPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state OnCallPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	input := modelToInput(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateOnCallPolicy(id, input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating on-call policy", fmt.Sprintf("Could not update on-call policy %d: %s", id, err))
		return
	}

	// If is_default changed to true, set it.
	if plan.IsDefault.ValueBool() && !result.IsDefault {
		result, err = r.client.SetDefaultOnCallPolicy(result.ID)
		if err != nil {
			resp.Diagnostics.AddError("Error setting default on-call policy", fmt.Sprintf("Policy updated but could not set as default: %s", err))
			return
		}
	}

	apiToModel(result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated on-call policy", map[string]any{"id": result.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *onCallPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OnCallPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	err := r.client.DeleteOnCallPolicy(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "on-call policy already deleted", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError("Error deleting on-call policy", fmt.Sprintf("Could not delete on-call policy %d: %s", id, err))
		return
	}

	tflog.Trace(ctx, "deleted on-call policy", map[string]any{"id": id})
}

func (r *onCallPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Could not parse on-call policy ID %q: %s", req.ID, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func modelToInput(ctx context.Context, m *OnCallPolicyModel, diags *diag.Diagnostics) apiclient.OnCallPolicyInput {
	scopeType := m.ScopeType.ValueString()
	input := apiclient.OnCallPolicyInput{
		Name:      m.Name.ValueString(),
		Enabled:   m.Enabled.ValueBool(),
		ScopeType: &scopeType,
	}

	input.MonitorGroupID = helpers.IntPtr(m.MonitorGroupID)

	// Build steps from nested blocks.
	if !m.Steps.IsNull() && !m.Steps.IsUnknown() {
		var stepModels []OnCallPolicyStepModel
		diags.Append(m.Steps.ElementsAs(ctx, &stepModels, false)...)
		if diags.HasError() {
			return input
		}

		steps := make([]apiclient.OnCallPolicyStepInput, len(stepModels))
		for i, sm := range stepModels {
			steps[i] = apiclient.OnCallPolicyStepInput{
				TargetType:   sm.TargetType.ValueString(),
				DelayMinutes: int(sm.DelayMinutes.ValueInt64()),
			}
			if !sm.ScheduleID.IsNull() && !sm.ScheduleID.IsUnknown() {
				id := int(sm.ScheduleID.ValueInt64())
				steps[i].ScheduleID = &id
			}
			if !sm.OrganizationMemberID.IsNull() && !sm.OrganizationMemberID.IsUnknown() {
				id := int(sm.OrganizationMemberID.ValueInt64())
				steps[i].OrganizationMemberID = &id
			}
		}
		input.Steps = steps
	}

	return input
}

func apiToModel(p *apiclient.OnCallPolicy, m *OnCallPolicyModel, diags *diag.Diagnostics) {
	m.ID = types.Int64Value(int64(p.ID))
	m.Name = types.StringValue(p.Name)
	m.Enabled = types.BoolValue(p.Enabled)
	m.IsDefault = types.BoolValue(p.IsDefault)
	m.ScopeType = types.StringValue(p.ScopeType)

	if p.MonitorGroup != nil {
		m.MonitorGroupID = types.Int64Value(int64(p.MonitorGroup.ID))
	} else {
		m.MonitorGroupID = types.Int64Null()
	}

	// Sort steps by position for stable ordering.
	sorted := make([]apiclient.OnCallPolicyStepRef, len(p.Steps))
	copy(sorted, p.Steps)
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
			objVal, d := types.ObjectValue(stepAttrTypes, attrs)
			diags.Append(d...)
			stepValues[i] = objVal
		}
		stepObjType := types.ObjectType{AttrTypes: stepAttrTypes}
		listVal, d := types.ListValue(stepObjType, stepValues)
		diags.Append(d...)
		m.Steps = listVal
	} else {
		stepObjType := types.ObjectType{AttrTypes: stepAttrTypes}
		m.Steps = types.ListNull(stepObjType)
	}
}
