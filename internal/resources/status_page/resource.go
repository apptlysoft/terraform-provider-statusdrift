package status_page

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &statusPageResource{}
	_ resource.ResourceWithImportState = &statusPageResource{}
)

type statusPageResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &statusPageResource{}
}

func (r *statusPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *statusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the status page.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The URL slug of the status page.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the status page.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of status page. Valid values: all, monitor, group, tag, custom.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("all", "monitor", "group", "tag", "custom"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the status page is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"description": schema.StringAttribute{
				Description: "A description of the status page.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(2000),
				},
			},
			"monitor_id": schema.Int64Attribute{
				Description: "The monitor ID. Required when type is 'monitor'.",
				Optional:    true,
			},
			"monitor_group_id": schema.Int64Attribute{
				Description: "The monitor group ID. Required when type is 'group'.",
				Optional:    true,
			},
			"tag_id": schema.Int64Attribute{
				Description: "The tag ID. Required when type is 'tag'.",
				Optional:    true,
			},
			"show_announcements": schema.BoolAttribute{
				Description: "Whether to show announcements on the status page.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"show_incidents": schema.BoolAttribute{
				Description: "Whether to show incidents on the status page.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"incident_display_types": schema.ListAttribute{
				Description: "List of incident types to display. Valid values: all, down, warning, manual.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidatorStringValues("all", "down", "warning", "manual"),
				},
			},
			"show_history": schema.BoolAttribute{
				Description: "Whether to show history on the status page.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"history_days": schema.Int64Attribute{
				Description: "The number of days of history to display (1-365).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(90),
				Validators: []validator.Int64{
					int64validator.Between(1, 365),
				},
			},
			"show_uptime": schema.BoolAttribute{
				Description: "Whether to show uptime percentage on the status page.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"show_response_time": schema.BoolAttribute{
				Description: "Whether to show response time on the status page.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"show_overall_status": schema.BoolAttribute{
				Description: "Whether to show overall status on the status page.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"logo_url": schema.StringAttribute{
				Description: "URL of the logo to display on the status page.",
				Optional:    true,
			},
			"favicon_url": schema.StringAttribute{
				Description: "URL of the favicon for the status page.",
				Optional:    true,
			},
			"custom_domain": schema.StringAttribute{
				Description: "Custom domain for the status page.",
				Optional:    true,
			},
			"primary_color": schema.StringAttribute{
				Description: "Primary color in hex format (e.g. '#FF5733').",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexpHexColor(),
						"must be a valid hex color (e.g. '#FF5733' or '#fff')",
					),
				},
			},
			"custom_css": schema.StringAttribute{
				Description: "Custom CSS for the status page.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(65535),
				},
			},
			"hide_powered_by": schema.BoolAttribute{
				Description: "Whether to hide the 'Powered by StatusDrift' branding.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"password": schema.StringAttribute{
				Description: "Password to protect the status page.",
				Optional:    true,
				Sensitive:   true,
			},
			"timezone": schema.StringAttribute{
				Description: "The timezone for the status page.",
				Optional:    true,
			},
			"google_analytics_id": schema.StringAttribute{
				Description: "Google Analytics measurement ID (e.g. 'G-XXXXXXXXXX').",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexpGoogleAnalytics(),
						"must be a valid Google Analytics ID (e.g. 'G-ABC123')",
					),
				},
			},
			"color_theme": schema.StringAttribute{
				Description: "The color theme. Valid values: light, dark.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("light"),
				Validators: []validator.String{
					stringvalidator.OneOf("light", "dark"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"social_links": schema.ListNestedBlock{
				Description: "Social media links to display on the status page.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"platform": schema.StringAttribute{
							Description: "The social media platform. Valid values: twitter, facebook, linkedin, github, instagram, youtube, discord, slack.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("twitter", "facebook", "linkedin", "github", "instagram", "youtube", "discord", "slack"),
							},
						},
						"url": schema.StringAttribute{
							Description: "The URL of the social media profile or page.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *statusPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *statusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := modelToInput(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating status page", map[string]interface{}{
		"name": input.Name,
		"type": input.Type,
	})

	statusPage, err := r.client.CreateStatusPage(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Status Page",
			fmt.Sprintf("Could not create status page: %s", err),
		)
		return
	}

	diags = apiToModel(ctx, statusPage, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	tflog.Debug(ctx, "Reading status page", map[string]interface{}{
		"id": id,
	})

	statusPage, err := r.client.GetStatusPage(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "Status page not found, removing from state", map[string]interface{}{
				"id": id,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Status Page",
			fmt.Sprintf("Could not read status page %d: %s", id, err),
		)
		return
	}

	diags := apiToModel(ctx, statusPage, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *statusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StatusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(plan.ID.ValueInt64())

	input, diags := modelToInput(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating status page", map[string]interface{}{
		"id": id,
	})

	statusPage, err := r.client.UpdateStatusPage(id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Status Page",
			fmt.Sprintf("Could not update status page %d: %s", id, err),
		)
		return
	}

	diags = apiToModel(ctx, statusPage, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	tflog.Debug(ctx, "Deleting status page", map[string]interface{}{
		"id": id,
	})

	err := r.client.DeleteStatusPage(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Status Page",
			fmt.Sprintf("Could not delete status page %d: %s", id, err),
		)
	}
}

func (r *statusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("The status page ID must be a numeric value, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// ---------------------------------------------------------------------------
// Model-to-API conversion
// ---------------------------------------------------------------------------

func modelToInput(ctx context.Context, m *StatusPageModel) (apiclient.StatusPageInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := apiclient.StatusPageInput{
		Name:              m.Name.ValueString(),
		Type:              m.Type.ValueString(),
		Enabled:           m.Enabled.ValueBool(),
		ShowAnnouncements: m.ShowAnnouncements.ValueBool(),
		ShowIncidents:     m.ShowIncidents.ValueBool(),
		ShowHistory:       m.ShowHistory.ValueBool(),
		HistoryDays:       int(m.HistoryDays.ValueInt64()),
		ShowUptime:        m.ShowUptime.ValueBool(),
		ShowResponseTime:  m.ShowResponseTime.ValueBool(),
		ShowOverallStatus: m.ShowOverallStatus.ValueBool(),
		HidePoweredBy:     m.HidePoweredBy.ValueBool(),
		ColorTheme:        m.ColorTheme.ValueString(),
	}

	input.Description = helpers.StringPtr(m.Description)

	input.MonitorID = helpers.IntPtr(m.MonitorID)

	input.MonitorGroupID = helpers.IntPtr(m.MonitorGroupID)

	input.TagID = helpers.IntPtr(m.TagID)

	if !m.IncidentDisplayTypes.IsNull() && !m.IncidentDisplayTypes.IsUnknown() {
		var displayTypes []string
		diags.Append(m.IncidentDisplayTypes.ElementsAs(ctx, &displayTypes, false)...)
		input.IncidentDisplayTypes = displayTypes
	}

	input.LogoURL = helpers.StringPtr(m.LogoURL)

	input.FaviconURL = helpers.StringPtr(m.FaviconURL)

	input.CustomDomain = helpers.StringPtr(m.CustomDomain)

	input.PrimaryColor = helpers.StringPtr(m.PrimaryColor)

	input.CustomCSS = helpers.StringPtr(m.CustomCSS)

	input.Password = helpers.StringPtr(m.Password)

	input.Timezone = helpers.StringPtr(m.Timezone)

	input.GoogleAnalyticsID = helpers.StringPtr(m.GoogleAnalyticsID)

	if !m.SocialLinks.IsNull() && !m.SocialLinks.IsUnknown() {
		var socialLinkModels []SocialLinkModel
		diags.Append(m.SocialLinks.ElementsAs(ctx, &socialLinkModels, false)...)
		if !diags.HasError() {
			links := make([]apiclient.SocialLink, len(socialLinkModels))
			for i, sl := range socialLinkModels {
				links[i] = apiclient.SocialLink{
					Platform: sl.Platform.ValueString(),
					URL:      sl.URL.ValueString(),
				}
			}
			input.SocialLinks = links
		}
	}

	return input, diags
}

// ---------------------------------------------------------------------------
// API-to-Model conversion
// ---------------------------------------------------------------------------

func apiToModel(ctx context.Context, sp *apiclient.StatusPage, m *StatusPageModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.Int64Value(int64(sp.ID))
	m.Slug = types.StringValue(sp.Slug)
	m.Name = types.StringValue(sp.Name)
	m.Type = types.StringValue(sp.Type)
	m.Enabled = types.BoolValue(sp.Enabled)
	m.ShowAnnouncements = types.BoolValue(sp.ShowAnnouncements)
	m.ShowIncidents = types.BoolValue(sp.ShowIncidents)
	m.ShowHistory = types.BoolValue(sp.ShowHistory)
	m.HistoryDays = types.Int64Value(int64(sp.HistoryDays))
	m.ShowUptime = types.BoolValue(sp.ShowUptime)
	m.ShowResponseTime = types.BoolValue(sp.ShowResponseTime)
	m.ShowOverallStatus = types.BoolValue(sp.ShowOverallStatus)
	m.HidePoweredBy = types.BoolValue(sp.HidePoweredBy)
	m.ColorTheme = types.StringValue(sp.ColorTheme)

	if sp.Description != nil {
		m.Description = types.StringValue(*sp.Description)
	} else {
		m.Description = types.StringNull()
	}

	if sp.MonitorID != nil {
		m.MonitorID = types.Int64Value(int64(*sp.MonitorID))
	} else {
		m.MonitorID = types.Int64Null()
	}

	if sp.MonitorGroupID != nil {
		m.MonitorGroupID = types.Int64Value(int64(*sp.MonitorGroupID))
	} else {
		m.MonitorGroupID = types.Int64Null()
	}

	if sp.TagID != nil {
		m.TagID = types.Int64Value(int64(*sp.TagID))
	} else {
		m.TagID = types.Int64Null()
	}

	// IncidentDisplayTypes — the API may return a default (e.g. ["all"]) even
	// when the user didn't set it. Preserve null to avoid spurious diffs.
	if len(sp.IncidentDisplayTypes) > 0 {
		if m.IncidentDisplayTypes.IsNull() {
			// User didn't set it; keep null
		} else {
			displayTypes, d := types.ListValueFrom(ctx, types.StringType, sp.IncidentDisplayTypes)
			diags.Append(d...)
			m.IncidentDisplayTypes = displayTypes
		}
	} else {
		m.IncidentDisplayTypes = types.ListNull(types.StringType)
	}

	if sp.LogoURL != nil {
		m.LogoURL = types.StringValue(*sp.LogoURL)
	} else {
		m.LogoURL = types.StringNull()
	}

	if sp.FaviconURL != nil {
		m.FaviconURL = types.StringValue(*sp.FaviconURL)
	} else {
		m.FaviconURL = types.StringNull()
	}

	if sp.CustomDomain != nil {
		m.CustomDomain = types.StringValue(*sp.CustomDomain)
	} else {
		m.CustomDomain = types.StringNull()
	}

	if sp.PrimaryColor != nil {
		m.PrimaryColor = types.StringValue(*sp.PrimaryColor)
	} else {
		m.PrimaryColor = types.StringNull()
	}

	if sp.CustomCSS != nil {
		m.CustomCSS = types.StringValue(*sp.CustomCSS)
	} else {
		m.CustomCSS = types.StringNull()
	}

	// Password — the API never returns the real value; always preserve plan/state.
	// m.Password already holds the plan (Create/Update) or prior state (Read).

	if sp.Timezone != nil {
		m.Timezone = types.StringValue(*sp.Timezone)
	} else {
		m.Timezone = types.StringNull()
	}

	if sp.GoogleAnalyticsID != nil {
		m.GoogleAnalyticsID = types.StringValue(*sp.GoogleAnalyticsID)
	} else {
		m.GoogleAnalyticsID = types.StringNull()
	}

	if len(sp.SocialLinks) > 0 {
		socialLinkObjType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"platform": types.StringType,
				"url":      types.StringType,
			},
		}

		socialLinkValues := make([]attr.Value, len(sp.SocialLinks))
		for i, sl := range sp.SocialLinks {
			obj, d := types.ObjectValue(
				socialLinkObjType.AttrTypes,
				map[string]attr.Value{
					"platform": types.StringValue(sl.Platform),
					"url":      types.StringValue(sl.URL),
				},
			)
			diags.Append(d...)
			socialLinkValues[i] = obj
		}

		socialLinksList, d := types.ListValue(socialLinkObjType, socialLinkValues)
		diags.Append(d...)
		m.SocialLinks = socialLinksList
	} else {
		socialLinkObjType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"platform": types.StringType,
				"url":      types.StringType,
			},
		}
		m.SocialLinks = types.ListNull(socialLinkObjType)
	}

	return diags
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func regexpHexColor() *regexp.Regexp {
	return regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
}

func regexpGoogleAnalytics() *regexp.Regexp {
	return regexp.MustCompile(`^G-[A-Z0-9]+$`)
}

// listvalidatorStringValues returns a list validator that checks each element
// is one of the given allowed values.
func listvalidatorStringValues(allowed ...string) validator.List {
	return listvalidator.ValueStringsAre(
		stringvalidator.OneOf(allowed...),
	)
}
