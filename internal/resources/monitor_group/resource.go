package monitor_group

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var (
	_ resource.Resource                = &monitorGroupResource{}
	_ resource.ResourceWithImportState = &monitorGroupResource{}
)

type monitorGroupResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &monitorGroupResource{}
}

func (r *monitorGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_group"
}

func (r *monitorGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift monitor group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the monitor group.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the monitor group (max 255 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
		},
	}
}

func (r *monitorGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *monitorGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := apiclient.MonitorGroupInput{
		Name: plan.Name.ValueString(),
	}

	result, err := r.client.CreateMonitorGroup(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating monitor group",
			fmt.Sprintf("Could not create monitor group: %s", err),
		)
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.Name = types.StringValue(result.Name)

	tflog.Trace(ctx, "created monitor group", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *monitorGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MonitorGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	result, err := r.client.GetMonitorGroup(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "monitor group not found, removing from state", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading monitor group",
			fmt.Sprintf("Could not read monitor group %d: %s", id, err),
		)
		return
	}

	state.ID = types.Int64Value(int64(result.ID))
	state.Name = types.StringValue(result.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *monitorGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state MonitorGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	input := apiclient.MonitorGroupInput{
		Name: plan.Name.ValueString(),
	}

	result, err := r.client.UpdateMonitorGroup(id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating monitor group",
			fmt.Sprintf("Could not update monitor group %d: %s", id, err),
		)
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.Name = types.StringValue(result.Name)

	tflog.Trace(ctx, "updated monitor group", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *monitorGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MonitorGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	err := r.client.DeleteMonitorGroup(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "monitor group already deleted", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting monitor group",
			fmt.Sprintf("Could not delete monitor group %d: %s", id, err),
		)
		return
	}

	tflog.Trace(ctx, "deleted monitor group", map[string]any{"id": id})
}

func (r *monitorGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Could not parse monitor group ID %q: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
