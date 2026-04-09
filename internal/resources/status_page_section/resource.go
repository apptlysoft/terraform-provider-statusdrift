package status_page_section

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = &statusPageSectionResource{}
	_ resource.ResourceWithImportState = &statusPageSectionResource{}
)

type statusPageSectionResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &statusPageSectionResource{}
}

func (r *statusPageSectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page_section"
}

func (r *statusPageSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift status page section.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the status page section.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status_page_id": schema.Int64Attribute{
				Description: "The ID of the status page this section belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the section.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the section.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(2000),
				},
			},
			"position": schema.Int64Attribute{
				Description: "The display position of the section (0 or greater).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"expanded": schema.BoolAttribute{
				Description: "Whether the section is expanded by default.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the section is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *statusPageSectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *statusPageSectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusPageSectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(plan.StatusPageID.ValueInt64())
	input := modelToInput(&plan)

	tflog.Debug(ctx, "Creating status page section", map[string]interface{}{
		"status_page_id": statusPageID,
		"name":           input.Name,
	})

	section, err := r.client.CreateStatusPageSection(statusPageID, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Status Page Section",
			fmt.Sprintf("Could not create status page section: %s", err),
		)
		return
	}

	apiToModel(section, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageSectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusPageSectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(state.StatusPageID.ValueInt64())
	id := int(state.ID.ValueInt64())

	tflog.Debug(ctx, "Reading status page section", map[string]interface{}{
		"status_page_id": statusPageID,
		"id":             id,
	})

	section, err := r.client.GetStatusPageSection(statusPageID, id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "Status page section not found, removing from state", map[string]interface{}{
				"status_page_id": statusPageID,
				"id":             id,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Status Page Section",
			fmt.Sprintf("Could not read status page section %d: %s", id, err),
		)
		return
	}

	apiToModel(section, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *statusPageSectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StatusPageSectionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(plan.StatusPageID.ValueInt64())
	id := int(plan.ID.ValueInt64())
	input := modelToInput(&plan)

	tflog.Debug(ctx, "Updating status page section", map[string]interface{}{
		"status_page_id": statusPageID,
		"id":             id,
	})

	section, err := r.client.UpdateStatusPageSection(statusPageID, id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Status Page Section",
			fmt.Sprintf("Could not update status page section %d: %s", id, err),
		)
		return
	}

	apiToModel(section, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageSectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusPageSectionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(state.StatusPageID.ValueInt64())
	id := int(state.ID.ValueInt64())

	tflog.Debug(ctx, "Deleting status page section", map[string]interface{}{
		"status_page_id": statusPageID,
		"id":             id,
	})

	err := r.client.DeleteStatusPageSection(statusPageID, id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Status Page Section",
			fmt.Sprintf("Could not delete status page section %d: %s", id, err),
		)
	}
}

func (r *statusPageSectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format '{status_page_id}/{section_id}', got: %q", req.ID),
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

	sectionID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse section_id %q as integer: %s", parts[1], err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("status_page_id"), statusPageID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), sectionID)...)
}

// ---------------------------------------------------------------------------
// Model-to-API conversion
// ---------------------------------------------------------------------------

func modelToInput(m *StatusPageSectionModel) apiclient.StatusPageSectionInput {
	input := apiclient.StatusPageSectionInput{
		Name:     m.Name.ValueString(),
		Expanded: m.Expanded.ValueBool(),
		Enabled:  m.Enabled.ValueBool(),
	}

	input.Description = helpers.StringPtr(m.Description)
	input.Position = helpers.IntPtr(m.Position)

	return input
}

// ---------------------------------------------------------------------------
// API-to-Model conversion
// ---------------------------------------------------------------------------

func apiToModel(section *apiclient.StatusPageSection, m *StatusPageSectionModel) {
	m.ID = types.Int64Value(int64(section.ID))
	m.Name = types.StringValue(section.Name)
	m.Expanded = types.BoolValue(section.Expanded)
	m.Enabled = types.BoolValue(section.Enabled)

	if section.Description != nil {
		m.Description = types.StringValue(*section.Description)
	} else {
		m.Description = types.StringNull()
	}

	if section.Position != nil {
		m.Position = types.Int64Value(int64(*section.Position))
	} else {
		m.Position = types.Int64Null()
	}
}
