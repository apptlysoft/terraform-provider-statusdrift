package status_page

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &statusPageDataSource{}

type statusPageDataSource struct {
	client *apiclient.Client
}

type statusPageDataSourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Slug              types.String `tfsdk:"slug"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	Description       types.String `tfsdk:"description"`
	MonitorID         types.Int64  `tfsdk:"monitor_id"`
	MonitorGroupID    types.Int64  `tfsdk:"monitor_group_id"`
	TagID             types.Int64  `tfsdk:"tag_id"`
	ShowAnnouncements types.Bool   `tfsdk:"show_announcements"`
	ShowIncidents     types.Bool   `tfsdk:"show_incidents"`
	ShowHistory       types.Bool   `tfsdk:"show_history"`
	HistoryDays       types.Int64  `tfsdk:"history_days"`
	ShowUptime        types.Bool   `tfsdk:"show_uptime"`
	ShowResponseTime  types.Bool   `tfsdk:"show_response_time"`
	ShowOverallStatus types.Bool   `tfsdk:"show_overall_status"`
	CustomDomain      types.String `tfsdk:"custom_domain"`
	HidePoweredBy     types.Bool   `tfsdk:"hide_powered_by"`
	Timezone          types.String `tfsdk:"timezone"`
	ColorTheme        types.String `tfsdk:"color_theme"`
}

func NewDataSource() datasource.DataSource {
	return &statusPageDataSource{}
}

func (d *statusPageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (d *statusPageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single status page by ID, name, or slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the status page.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name"), path.MatchRoot("slug")),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The URL slug of the status page.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the status page.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the status page.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the status page is enabled.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the status page.",
				Computed:    true,
			},
			"monitor_id": schema.Int64Attribute{
				Description: "The monitor ID associated with this status page.",
				Computed:    true,
			},
			"monitor_group_id": schema.Int64Attribute{
				Description: "The monitor group ID associated with this status page.",
				Computed:    true,
			},
			"tag_id": schema.Int64Attribute{
				Description: "The tag ID associated with this status page.",
				Computed:    true,
			},
			"show_announcements": schema.BoolAttribute{
				Description: "Whether to show announcements.",
				Computed:    true,
			},
			"show_incidents": schema.BoolAttribute{
				Description: "Whether to show incidents.",
				Computed:    true,
			},
			"show_history": schema.BoolAttribute{
				Description: "Whether to show history.",
				Computed:    true,
			},
			"history_days": schema.Int64Attribute{
				Description: "Number of days of history to show.",
				Computed:    true,
			},
			"show_uptime": schema.BoolAttribute{
				Description: "Whether to show uptime percentage.",
				Computed:    true,
			},
			"show_response_time": schema.BoolAttribute{
				Description: "Whether to show response time.",
				Computed:    true,
			},
			"show_overall_status": schema.BoolAttribute{
				Description: "Whether to show the overall status indicator.",
				Computed:    true,
			},
			"custom_domain": schema.StringAttribute{
				Description: "Custom domain for the status page.",
				Computed:    true,
			},
			"hide_powered_by": schema.BoolAttribute{
				Description: "Whether to hide the powered-by branding.",
				Computed:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone for the status page.",
				Computed:    true,
			},
			"color_theme": schema.StringAttribute{
				Description: "Color theme of the status page.",
				Computed:    true,
			},
		},
	}
}

func (d *statusPageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *statusPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state statusPageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var page *apiclient.StatusPage

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		pages, err := d.client.ListStatusPagesWithSearch(name)
		if err != nil {
			resp.Diagnostics.AddError("Error searching status pages", err.Error())
			return
		}
		var matches []apiclient.StatusPage
		for _, p := range pages {
			if p.Name == name {
				matches = append(matches, p)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No status page found", fmt.Sprintf("No status page found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple status pages found", fmt.Sprintf("Multiple status pages found with name %q, use id instead.", name))
			return
		}
		page = &matches[0]
	} else if !state.Slug.IsNull() {
		slug := state.Slug.ValueString()
		pages, err := d.client.ListStatusPagesWithSearch(slug)
		if err != nil {
			resp.Diagnostics.AddError("Error searching status pages", err.Error())
			return
		}
		var matches []apiclient.StatusPage
		for _, p := range pages {
			if p.Slug == slug {
				matches = append(matches, p)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No status page found", fmt.Sprintf("No status page found with slug %q.", slug))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple status pages found", fmt.Sprintf("Multiple status pages found with slug %q, use id instead.", slug))
			return
		}
		page = &matches[0]
	} else {
		result, err := d.client.GetStatusPage(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading status page", err.Error())
			return
		}
		page = result
	}

	state.ID = types.Int64Value(int64(page.ID))
	state.Slug = types.StringValue(page.Slug)
	state.Name = types.StringValue(page.Name)
	state.Type = types.StringValue(page.Type)
	state.Enabled = types.BoolValue(page.Enabled)
	state.ShowAnnouncements = types.BoolValue(page.ShowAnnouncements)
	state.ShowIncidents = types.BoolValue(page.ShowIncidents)
	state.ShowHistory = types.BoolValue(page.ShowHistory)
	state.HistoryDays = types.Int64Value(int64(page.HistoryDays))
	state.ShowUptime = types.BoolValue(page.ShowUptime)
	state.ShowResponseTime = types.BoolValue(page.ShowResponseTime)
	state.ShowOverallStatus = types.BoolValue(page.ShowOverallStatus)
	state.HidePoweredBy = types.BoolValue(page.HidePoweredBy)
	state.ColorTheme = types.StringValue(page.ColorTheme)

	// Optional pointer fields
	if page.Description != nil {
		state.Description = types.StringValue(*page.Description)
	} else {
		state.Description = types.StringNull()
	}
	if page.MonitorID != nil {
		state.MonitorID = types.Int64Value(int64(*page.MonitorID))
	} else {
		state.MonitorID = types.Int64Null()
	}
	if page.MonitorGroupID != nil {
		state.MonitorGroupID = types.Int64Value(int64(*page.MonitorGroupID))
	} else {
		state.MonitorGroupID = types.Int64Null()
	}
	if page.TagID != nil {
		state.TagID = types.Int64Value(int64(*page.TagID))
	} else {
		state.TagID = types.Int64Null()
	}
	if page.CustomDomain != nil {
		state.CustomDomain = types.StringValue(*page.CustomDomain)
	} else {
		state.CustomDomain = types.StringNull()
	}
	if page.Timezone != nil {
		state.Timezone = types.StringValue(*page.Timezone)
	} else {
		state.Timezone = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
