package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	datasourceAlert "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/alert"
	datasourceAlerts "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/alerts"
	datasourceAPIKey "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/api_key"
	datasourceAPIKeys "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/api_keys"
	datasourceCollaborator "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/collaborator"
	datasourceCollaborators "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/collaborators"
	datasourceIncident "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/incident"
	datasourceIncidents "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/incidents"
	datasourceLocations "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/locations"
	datasourceMaintenanceWindow "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/maintenance_window"
	datasourceMaintenanceWindows "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/maintenance_windows"
	datasourceMonitor "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/monitor"
	datasourceMonitorGroup "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/monitor_group"
	datasourceMonitorGroups "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/monitor_groups"
	datasourceMonitors "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/monitors"
	datasourceNotificationChannel "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/notification_channel"
	datasourceNotificationChannels "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/notification_channels"
	datasourceOnCallPolicies "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/on_call_policies"
	datasourceOnCallPolicy "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/on_call_policy"
	datasourceOnCallSchedule "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/on_call_schedule"
	datasourceOnCallSchedules "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/on_call_schedules"
	datasourceSlaPolicies "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/sla_policies"
	datasourceSlaPolicy "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/sla_policy"
	datasourceStatusPage "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/status_page"
	datasourceStatusPageAnnouncement "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/status_page_announcement"
	datasourceStatusPageAnnouncements "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/status_page_announcements"
	datasourceStatusPageSection "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/status_page_section"
	datasourceStatusPageSections "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/status_page_sections"
	datasourceStatusPages "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/status_pages"
	datasourceTag "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/tag"
	datasourceTags "github.com/apptlysoft/terraform-provider-statusdrift/internal/datasources/tags"
	resourceAlert "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/alert"
	resourceAPIKey "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/api_key"
	resourceIncident "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/incident"
	resourceMaintenanceWindow "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/maintenance_window"
	resourceMonitor "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/monitor"
	resourceMonitorGroup "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/monitor_group"
	resourceNotificationChannel "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/notification_channel"
	resourceOnCallPolicy "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/on_call_policy"
	resourceOnCallSchedule "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/on_call_schedule"
	resourceOnCallScheduleOverride "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/on_call_schedule_override"
	resourceSlaPolicy "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/sla_policy"
	resourceStatusPage "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/status_page"
	resourceStatusPageAnnouncement "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/status_page_announcement"
	resourceStatusPageSection "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/status_page_section"
	resourceTag "github.com/apptlysoft/terraform-provider-statusdrift/internal/resources/tag"
)

var _ provider.Provider = &StatusDriftProvider{}

type StatusDriftProvider struct {
	version string
}

type StatusDriftProviderModel struct {
	APIURL types.String `tfsdk:"api_url"`
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StatusDriftProvider{version: version}
	}
}

func (p *StatusDriftProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "statusdrift"
	resp.Version = p.version
}

func (p *StatusDriftProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing StatusDrift monitoring resources.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Description: "StatusDrift API URL. Can also be set via STATUSDRIFT_API_URL environment variable. Defaults to https://api.statusdrift.com.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "StatusDrift API key (must start with sdk_). Can also be set via STATUSDRIFT_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *StatusDriftProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config StatusDriftProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiURL := "https://api.statusdrift.com"
	if !config.APIURL.IsNull() && !config.APIURL.IsUnknown() {
		apiURL = config.APIURL.ValueString()
	} else if v := os.Getenv("STATUSDRIFT_API_URL"); v != "" {
		apiURL = v
	}

	var apiKey string
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	} else if v := os.Getenv("STATUSDRIFT_API_KEY"); v != "" {
		apiKey = v
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The StatusDrift API key must be set in the provider configuration or via the STATUSDRIFT_API_KEY environment variable.",
		)
		return
	}

	if !strings.HasPrefix(apiKey, "sdk_") {
		resp.Diagnostics.AddError(
			"Invalid API Key",
			"The StatusDrift API key must start with the 'sdk_' prefix.",
		)
		return
	}

	client := apiclient.NewClient(apiURL, apiKey, p.version)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *StatusDriftProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resourceMonitorGroup.NewResource,
		resourceTag.NewResource,
		resourceNotificationChannel.NewResource,
		resourceMonitor.NewResource,
		resourceAlert.NewResource,
		resourceStatusPage.NewResource,
		resourceStatusPageSection.NewResource,
		resourceStatusPageAnnouncement.NewResource,
		resourceMaintenanceWindow.NewResource,
		resourceIncident.NewResource,
		resourceAPIKey.NewResource,
		resourceSlaPolicy.NewResource,
		resourceOnCallSchedule.NewResource,
		resourceOnCallScheduleOverride.NewResource,
		resourceOnCallPolicy.NewResource,
	}
}

func (p *StatusDriftProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasourceMonitor.NewDataSource,
		datasourceMonitors.NewDataSource,
		datasourceMonitorGroup.NewDataSource,
		datasourceMonitorGroups.NewDataSource,
		datasourceTag.NewDataSource,
		datasourceTags.NewDataSource,
		datasourceNotificationChannel.NewDataSource,
		datasourceNotificationChannels.NewDataSource,
		datasourceStatusPage.NewDataSource,
		datasourceStatusPages.NewDataSource,
		datasourceAlert.NewDataSource,
		datasourceAlerts.NewDataSource,
		datasourceIncident.NewDataSource,
		datasourceIncidents.NewDataSource,
		datasourceMaintenanceWindow.NewDataSource,
		datasourceMaintenanceWindows.NewDataSource,
		datasourceLocations.NewDataSource,
		datasourceStatusPageSection.NewDataSource,
		datasourceStatusPageSections.NewDataSource,
		datasourceStatusPageAnnouncement.NewDataSource,
		datasourceStatusPageAnnouncements.NewDataSource,
		datasourceAPIKey.NewDataSource,
		datasourceAPIKeys.NewDataSource,
		datasourceSlaPolicy.NewDataSource,
		datasourceSlaPolicies.NewDataSource,
		datasourceOnCallSchedule.NewDataSource,
		datasourceOnCallSchedules.NewDataSource,
		datasourceOnCallPolicy.NewDataSource,
		datasourceOnCallPolicies.NewDataSource,
		datasourceCollaborator.NewDataSource,
		datasourceCollaborators.NewDataSource,
	}
}
