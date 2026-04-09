package status_page_announcement

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &statusPageAnnouncementResource{}
	_ resource.ResourceWithImportState = &statusPageAnnouncementResource{}
)

type statusPageAnnouncementResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &statusPageAnnouncementResource{}
}

func (r *statusPageAnnouncementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page_announcement"
}

func (r *statusPageAnnouncementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift status page announcement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the announcement.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status_page_id": schema.Int64Attribute{
				Description: "The ID of the status page this announcement belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title of the announcement.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"content": schema.StringAttribute{
				Description: "The content of the announcement.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(10000),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of announcement. Valid values: info, warning, critical, maintenance, success.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("info", "warning", "critical", "maintenance", "success"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the announcement is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"starts_at": schema.StringAttribute{
				Description: "The start date and time in ISO 8601 format.",
				Optional:    true,
			},
			"ends_at": schema.StringAttribute{
				Description: "The end date and time in ISO 8601 format.",
				Optional:    true,
			},
		},
	}
}

func (r *statusPageAnnouncementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *statusPageAnnouncementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusPageAnnouncementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(plan.StatusPageID.ValueInt64())
	input := modelToInput(&plan)

	tflog.Debug(ctx, "Creating status page announcement", map[string]interface{}{
		"status_page_id": statusPageID,
		"title":          input.Title,
	})

	announcement, err := r.client.CreateStatusPageAnnouncement(statusPageID, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Status Page Announcement",
			fmt.Sprintf("Could not create status page announcement: %s", err),
		)
		return
	}

	apiToModel(announcement, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageAnnouncementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusPageAnnouncementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(state.StatusPageID.ValueInt64())
	id := int(state.ID.ValueInt64())

	tflog.Debug(ctx, "Reading status page announcement", map[string]interface{}{
		"status_page_id": statusPageID,
		"id":             id,
	})

	announcement, err := r.client.GetStatusPageAnnouncement(statusPageID, id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "Status page announcement not found, removing from state", map[string]interface{}{
				"status_page_id": statusPageID,
				"id":             id,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Status Page Announcement",
			fmt.Sprintf("Could not read status page announcement %d: %s", id, err),
		)
		return
	}

	apiToModel(announcement, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *statusPageAnnouncementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StatusPageAnnouncementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(plan.StatusPageID.ValueInt64())
	id := int(plan.ID.ValueInt64())
	input := modelToInput(&plan)

	tflog.Debug(ctx, "Updating status page announcement", map[string]interface{}{
		"status_page_id": statusPageID,
		"id":             id,
	})

	announcement, err := r.client.UpdateStatusPageAnnouncement(statusPageID, id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Status Page Announcement",
			fmt.Sprintf("Could not update status page announcement %d: %s", id, err),
		)
		return
	}

	apiToModel(announcement, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageAnnouncementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusPageAnnouncementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(state.StatusPageID.ValueInt64())
	id := int(state.ID.ValueInt64())

	tflog.Debug(ctx, "Deleting status page announcement", map[string]interface{}{
		"status_page_id": statusPageID,
		"id":             id,
	})

	err := r.client.DeleteStatusPageAnnouncement(statusPageID, id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Status Page Announcement",
			fmt.Sprintf("Could not delete status page announcement %d: %s", id, err),
		)
	}
}

func (r *statusPageAnnouncementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format '{status_page_id}/{announcement_id}', got: %q", req.ID),
		)
		return
	}

	statusPageID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse status_page_id %q as integer: %s", parts[0], err),
		)
		return
	}

	announcementID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse announcement_id %q as integer: %s", parts[1], err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("status_page_id"), statusPageID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), announcementID)...)
}

// ---------------------------------------------------------------------------
// Model-to-API conversion
// ---------------------------------------------------------------------------

func modelToInput(m *StatusPageAnnouncementModel) apiclient.StatusPageAnnouncementInput {
	input := apiclient.StatusPageAnnouncementInput{
		Title:   m.Title.ValueString(),
		Content: m.Content.ValueString(),
		Type:    m.Type.ValueString(),
		Enabled: m.Enabled.ValueBool(),
	}

	input.StartsAt = helpers.StringPtr(m.StartsAt)
	input.EndsAt = helpers.StringPtr(m.EndsAt)

	return input
}

// ---------------------------------------------------------------------------
// API-to-Model conversion
// ---------------------------------------------------------------------------

func apiToModel(a *apiclient.StatusPageAnnouncement, m *StatusPageAnnouncementModel) {
	m.ID = types.Int64Value(int64(a.ID))
	m.Title = types.StringValue(a.Title)
	m.Content = types.StringValue(a.Content)
	m.Type = types.StringValue(a.Type)
	m.Enabled = types.BoolValue(a.Enabled)

	if a.StartsAt != nil {
		m.StartsAt = types.StringValue(normalizeTimestamp(*a.StartsAt, m.StartsAt))
	} else {
		m.StartsAt = types.StringNull()
	}

	if a.EndsAt != nil {
		m.EndsAt = types.StringValue(normalizeTimestamp(*a.EndsAt, m.EndsAt))
	} else {
		m.EndsAt = types.StringNull()
	}
}

// normalizeTimestamp keeps the state/plan timestamp format when the API returns
// a semantically equivalent value (e.g. "Z" vs "+00:00").
func normalizeTimestamp(apiVal string, stateVal types.String) string {
	if stateVal.IsNull() || stateVal.IsUnknown() {
		return apiVal
	}
	apiTime, err1 := time.Parse(time.RFC3339, apiVal)
	stateTime, err2 := time.Parse(time.RFC3339, stateVal.ValueString())
	if err1 == nil && err2 == nil && apiTime.Equal(stateTime) {
		return stateVal.ValueString()
	}
	return apiVal
}
