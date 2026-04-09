package incident

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &incidentResource{}
	_ resource.ResourceWithImportState = &incidentResource{}
)

type incidentResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &incidentResource{}
}

func (r *incidentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incident"
}

func (r *incidentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift incident.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the incident.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"monitor_id": schema.Int64Attribute{
				Description: "The ID of the monitor this incident belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"message": schema.StringAttribute{
				Description: "The incident message.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(10000),
				},
			},
			"priority": schema.StringAttribute{
				Description: "The incident priority: low, medium, high, or critical.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("low", "medium", "high", "critical"),
				},
			},
			"resolved": schema.BoolAttribute{
				Description: "Whether the incident has been resolved. Set to true to resolve the incident.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *incidentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *incidentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IncidentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := apiclient.IncidentInput{
		MonitorID: int(plan.MonitorID.ValueInt64()),
	}

	input.Message = helpers.StringPtr(plan.Message)
	input.Priority = helpers.StringPtr(plan.Priority)

	incident, err := r.client.CreateIncident(input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating incident", err.Error())
		return
	}

	// Only take ID from the response; keep plan values for message/priority
	// since the API Create response may omit them.
	plan.ID = types.Int64Value(int64(incident.ID))

	// Create always returns an unresolved incident. If the plan wants it
	// resolved, call the resolve endpoint.
	if !plan.Resolved.IsNull() && plan.Resolved.ValueBool() {
		_, err := r.client.ResolveIncident(int(plan.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error resolving incident after create", err.Error())
			return
		}
		plan.Resolved = types.BoolValue(true)
	} else {
		plan.Resolved = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *incidentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IncidentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	incident, err := r.client.GetIncident(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading incident", err.Error())
		return
	}

	mapIncidentToState(incident, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *incidentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IncidentModel
	var state IncidentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	// Only resolved can change via update. When resolved transitions to true,
	// call the resolve endpoint.
	if !plan.Resolved.IsNull() && !plan.Resolved.IsUnknown() &&
		plan.Resolved.ValueBool() && !state.Resolved.ValueBool() {
		_, err := r.client.ResolveIncident(id)
		if err != nil {
			// Ignore "already resolved" — idempotent
			if !apiclient.IsBadRequest(err) {
				resp.Diagnostics.AddError("Error resolving incident", err.Error())
				return
			}
		}
	}

	// Keep plan values — the API responses may omit message/priority.
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *incidentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IncidentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	err := r.client.DeleteIncident(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting incident", err.Error())
	}
}

func (r *incidentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Could not parse ID %q as integer: %s", req.ID, err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(id))...)
}

func mapIncidentToState(incident *apiclient.Incident, state *IncidentModel) {
	state.ID = types.Int64Value(int64(incident.ID))
	state.MonitorID = types.Int64Value(int64(incident.MonitorID))
	state.Resolved = types.BoolValue(incident.Resolved || incident.ResolvedAt != nil)

	if incident.Message != nil {
		state.Message = types.StringValue(*incident.Message)
	}
	// else: API returned nil — preserve existing state value

	if incident.Priority != nil {
		state.Priority = types.StringValue(*incident.Priority)
	}
	// else: API returned nil — preserve existing state value
}
