package monitor

import "github.com/hashicorp/terraform-plugin-framework/types"

type MonitorModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Interval             types.Int64  `tfsdk:"interval"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	LocationType         types.String `tfsdk:"location_type"`
	LocationIDs          types.List   `tfsdk:"location_ids"`
	LocationNames        types.List   `tfsdk:"location_names"`
	Region               types.String `tfsdk:"region"`
	URL                  types.String `tfsdk:"url"`
	Host                 types.String `tfsdk:"host"`
	Port                 types.Int64  `tfsdk:"port"`
	HTTPMethod           types.String `tfsdk:"http_method"`
	HTTPPostBody         types.String `tfsdk:"http_post_body"`
	HTTPContentType      types.String `tfsdk:"http_content_type"`
	MaxRedirects         types.Int64  `tfsdk:"max_redirects"`
	AllowInvalidCert     types.Bool   `tfsdk:"allow_invalid_certificate"`
	CacheBuster          types.Bool   `tfsdk:"cache_buster"`
	HTTPAuth             types.Bool   `tfsdk:"http_auth"`
	HTTPUsername         types.String `tfsdk:"http_username"`
	HTTPPassword         types.String `tfsdk:"http_password"`
	RWTimeout            types.Int64  `tfsdk:"rw_timeout"`
	ConnectionTimeout    types.Int64  `tfsdk:"connection_timeout"`
	ExpectedStatusCode   types.String `tfsdk:"expected_status_code"`
	Keyword              types.String `tfsdk:"keyword"`
	KeywordNegation      types.Bool   `tfsdk:"keyword_negation"`
	MonitorGroupID       types.Int64  `tfsdk:"monitor_group_id"`
	TagIDs               types.List   `tfsdk:"tag_ids"`
	LocationsDownToAlert types.Int64  `tfsdk:"locations_down_to_alert"`
	ChecksDownToAlert    types.Int64  `tfsdk:"checks_down_to_alert"`
	WarningThresholdMs   types.Int64  `tfsdk:"warning_threshold_ms"`
	DNSHostname          types.String `tfsdk:"dns_hostname"`
	DNSServers           types.List   `tfsdk:"dns_servers"`
	DNSMonitoringMode    types.String `tfsdk:"dns_monitoring_mode"`
	Alerts               types.List   `tfsdk:"alert"`
}

type AlertModel struct {
	NotificationChannelIDs   types.List  `tfsdk:"notification_channel_ids"`
	Conditions               types.List  `tfsdk:"conditions"`
	DelayMinutes             types.Int64 `tfsdk:"delay_minutes"`
	Enabled                  types.Bool  `tfsdk:"enabled"`
	RecurringIntervalMinutes types.Int64 `tfsdk:"recurring_interval_minutes"`
	MaxRecurrences           types.Int64 `tfsdk:"max_recurrences"`
	NotifyOnCall             types.Bool  `tfsdk:"notify_on_call"`
	OnCallPolicyID           types.Int64 `tfsdk:"on_call_policy_id"`
}
