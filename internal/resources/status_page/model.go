package status_page

import "github.com/hashicorp/terraform-plugin-framework/types"

type StatusPageModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	Slug                 types.String `tfsdk:"slug"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	Description          types.String `tfsdk:"description"`
	MonitorID            types.Int64  `tfsdk:"monitor_id"`
	MonitorGroupID       types.Int64  `tfsdk:"monitor_group_id"`
	TagID                types.Int64  `tfsdk:"tag_id"`
	ShowAnnouncements    types.Bool   `tfsdk:"show_announcements"`
	ShowIncidents        types.Bool   `tfsdk:"show_incidents"`
	IncidentDisplayTypes types.List   `tfsdk:"incident_display_types"`
	ShowHistory          types.Bool   `tfsdk:"show_history"`
	HistoryDays          types.Int64  `tfsdk:"history_days"`
	ShowUptime           types.Bool   `tfsdk:"show_uptime"`
	ShowResponseTime     types.Bool   `tfsdk:"show_response_time"`
	ShowOverallStatus    types.Bool   `tfsdk:"show_overall_status"`
	LogoURL              types.String `tfsdk:"logo_url"`
	FaviconURL           types.String `tfsdk:"favicon_url"`
	CustomDomain         types.String `tfsdk:"custom_domain"`
	PrimaryColor         types.String `tfsdk:"primary_color"`
	CustomCSS            types.String `tfsdk:"custom_css"`
	HidePoweredBy        types.Bool   `tfsdk:"hide_powered_by"`
	Password             types.String `tfsdk:"password"`
	Timezone             types.String `tfsdk:"timezone"`
	GoogleAnalyticsID    types.String `tfsdk:"google_analytics_id"`
	SocialLinks          types.List   `tfsdk:"social_links"`
	ColorTheme           types.String `tfsdk:"color_theme"`
}

type SocialLinkModel struct {
	Platform types.String `tfsdk:"platform"`
	URL      types.String `tfsdk:"url"`
}
