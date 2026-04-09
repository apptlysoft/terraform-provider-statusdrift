package monitor

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

var _ datasource.DataSource = &monitorDataSource{}

type monitorDataSource struct {
	client *apiclient.Client
}

type monitorDataSourceModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Interval             types.Int64  `tfsdk:"interval"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	LocationType         types.String `tfsdk:"location_type"`
	LocationIDs          types.List   `tfsdk:"location_ids"`
	Region               types.String `tfsdk:"region"`
	URL                  types.String `tfsdk:"url"`
	Host                 types.String `tfsdk:"host"`
	Port                 types.Int64  `tfsdk:"port"`
	HTTPMethod           types.String `tfsdk:"http_method"`
	HTTPPostBody         types.String `tfsdk:"http_post_body"`
	HTTPContentType      types.String `tfsdk:"http_content_type"`
	MaxRedirects         types.Int64  `tfsdk:"max_redirects"`
	AllowInvalidCert     types.Bool   `tfsdk:"allow_invalid_cert"`
	CacheBuster          types.Bool   `tfsdk:"cache_buster"`
	HTTPAuth             types.Bool   `tfsdk:"http_auth"`
	HTTPUsername         types.String `tfsdk:"http_username"`
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
	DNSMonitoringMode    types.String `tfsdk:"dns_monitoring_mode"`
	Status               types.String `tfsdk:"status"`
	CreatedAt            types.String `tfsdk:"created_at"`
}

func NewDataSource() datasource.DataSource {
	return &monitorDataSource{}
}

func (d *monitorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (d *monitorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single monitor by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the monitor.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the monitor.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the monitor (http, ping, port, dns).",
				Computed:    true,
			},
			"interval": schema.Int64Attribute{
				Description: "Check interval in seconds.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the monitor is enabled.",
				Computed:    true,
			},
			"location_type": schema.StringAttribute{
				Description: "Location type (specific or region).",
				Computed:    true,
			},
			"location_ids": schema.ListAttribute{
				Description: "List of location IDs.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"region": schema.StringAttribute{
				Description: "Region for region-based location type.",
				Computed:    true,
			},
			"url": schema.StringAttribute{
				Description: "URL to monitor (for HTTP monitors).",
				Computed:    true,
			},
			"host": schema.StringAttribute{
				Description: "Host to monitor (for ping/port monitors).",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port to monitor (for port monitors).",
				Computed:    true,
			},
			"http_method": schema.StringAttribute{
				Description: "HTTP method.",
				Computed:    true,
			},
			"http_post_body": schema.StringAttribute{
				Description: "HTTP POST body.",
				Computed:    true,
			},
			"http_content_type": schema.StringAttribute{
				Description: "HTTP content type.",
				Computed:    true,
			},
			"max_redirects": schema.Int64Attribute{
				Description: "Maximum number of redirects to follow.",
				Computed:    true,
			},
			"allow_invalid_cert": schema.BoolAttribute{
				Description: "Whether to allow invalid SSL certificates.",
				Computed:    true,
			},
			"cache_buster": schema.BoolAttribute{
				Description: "Whether to append a cache-busting query parameter.",
				Computed:    true,
			},
			"http_auth": schema.BoolAttribute{
				Description: "Whether HTTP authentication is enabled.",
				Computed:    true,
			},
			"http_username": schema.StringAttribute{
				Description: "HTTP authentication username.",
				Computed:    true,
			},
			"rw_timeout": schema.Int64Attribute{
				Description: "Read/write timeout in milliseconds.",
				Computed:    true,
			},
			"connection_timeout": schema.Int64Attribute{
				Description: "Connection timeout in milliseconds.",
				Computed:    true,
			},
			"expected_status_code": schema.StringAttribute{
				Description: "Expected HTTP status code.",
				Computed:    true,
			},
			"keyword": schema.StringAttribute{
				Description: "Keyword to search for in the response.",
				Computed:    true,
			},
			"keyword_negation": schema.BoolAttribute{
				Description: "Whether keyword match is negated.",
				Computed:    true,
			},
			"monitor_group_id": schema.Int64Attribute{
				Description: "Monitor group ID.",
				Computed:    true,
			},
			"tag_ids": schema.ListAttribute{
				Description: "List of tag IDs.",
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"locations_down_to_alert": schema.Int64Attribute{
				Description: "Number of locations that must be down to trigger an alert.",
				Computed:    true,
			},
			"checks_down_to_alert": schema.Int64Attribute{
				Description: "Number of consecutive checks that must fail to trigger an alert.",
				Computed:    true,
			},
			"warning_threshold_ms": schema.Int64Attribute{
				Description: "Response time warning threshold in milliseconds.",
				Computed:    true,
			},
			"dns_hostname": schema.StringAttribute{
				Description: "DNS hostname to query.",
				Computed:    true,
			},
			"dns_monitoring_mode": schema.StringAttribute{
				Description: "DNS monitoring mode.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the monitor.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp.",
				Computed:    true,
			},
		},
	}
}

func (d *monitorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *monitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state monitorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var monitor *apiclient.Monitor

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		monitors, err := d.client.ListMonitorsWithSearch(name)
		if err != nil {
			resp.Diagnostics.AddError("Error searching monitors", err.Error())
			return
		}
		var matches []apiclient.Monitor
		for _, m := range monitors {
			if m.Name == name {
				matches = append(matches, m)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No monitor found", fmt.Sprintf("No monitor found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple monitors found", fmt.Sprintf("Multiple monitors found with name %q, use id instead.", name))
			return
		}
		monitor = &matches[0]
	} else {
		result, err := d.client.GetMonitor(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading monitor", err.Error())
			return
		}
		monitor = result
	}

	state.ID = types.Int64Value(int64(monitor.ID))
	state.Name = types.StringValue(monitor.Name)
	state.Type = types.StringValue(monitor.Type)
	state.Interval = types.Int64Value(int64(monitor.Interval))
	state.Enabled = types.BoolValue(monitor.Enabled)
	state.LocationType = types.StringValue(monitor.LocationType)
	state.AllowInvalidCert = types.BoolValue(monitor.AllowInvalidCert)
	state.CacheBuster = types.BoolValue(monitor.CacheBuster)
	state.HTTPAuth = types.BoolValue(monitor.HTTPAuth)
	state.KeywordNegation = types.BoolValue(monitor.KeywordNegation)
	state.LocationsDownToAlert = types.Int64Value(int64(monitor.LocationsDownToAlert))
	state.ChecksDownToAlert = types.Int64Value(int64(monitor.ChecksDownToAlert))
	state.WarningThresholdMs = types.Int64Value(int64(monitor.WarningThresholdMs))

	// Convert []int to types.List
	locationIDValues := make([]types.Int64, len(monitor.LocationIDs))
	for i, v := range monitor.LocationIDs {
		locationIDValues[i] = types.Int64Value(int64(v))
	}
	locationList, diags := types.ListValueFrom(ctx, types.Int64Type, monitor.LocationIDs)
	resp.Diagnostics.Append(diags...)
	state.LocationIDs = locationList

	tagList, diags := types.ListValueFrom(ctx, types.Int64Type, monitor.TagIDs)
	resp.Diagnostics.Append(diags...)
	state.TagIDs = tagList

	// Optional pointer fields
	if monitor.Region != nil {
		state.Region = types.StringValue(*monitor.Region)
	} else {
		state.Region = types.StringNull()
	}
	if monitor.URL != nil {
		state.URL = types.StringValue(*monitor.URL)
	} else {
		state.URL = types.StringNull()
	}
	if monitor.Host != nil {
		state.Host = types.StringValue(*monitor.Host)
	} else {
		state.Host = types.StringNull()
	}
	if monitor.Port != nil {
		state.Port = types.Int64Value(int64(*monitor.Port))
	} else {
		state.Port = types.Int64Null()
	}
	if monitor.HTTPMethod != nil {
		state.HTTPMethod = types.StringValue(*monitor.HTTPMethod)
	} else {
		state.HTTPMethod = types.StringNull()
	}
	if monitor.HTTPPostBody != nil {
		state.HTTPPostBody = types.StringValue(*monitor.HTTPPostBody)
	} else {
		state.HTTPPostBody = types.StringNull()
	}
	if monitor.HTTPContentType != nil {
		state.HTTPContentType = types.StringValue(*monitor.HTTPContentType)
	} else {
		state.HTTPContentType = types.StringNull()
	}
	if monitor.MaxRedirects != nil {
		state.MaxRedirects = types.Int64Value(int64(*monitor.MaxRedirects))
	} else {
		state.MaxRedirects = types.Int64Null()
	}
	if monitor.HTTPUsername != nil {
		state.HTTPUsername = types.StringValue(*monitor.HTTPUsername)
	} else {
		state.HTTPUsername = types.StringNull()
	}
	if monitor.RWTimeout != nil {
		state.RWTimeout = types.Int64Value(int64(*monitor.RWTimeout))
	} else {
		state.RWTimeout = types.Int64Null()
	}
	if monitor.ConnectionTimeout != nil {
		state.ConnectionTimeout = types.Int64Value(int64(*monitor.ConnectionTimeout))
	} else {
		state.ConnectionTimeout = types.Int64Null()
	}
	if monitor.ExpectedStatusCode != nil {
		state.ExpectedStatusCode = types.StringValue(*monitor.ExpectedStatusCode)
	} else {
		state.ExpectedStatusCode = types.StringNull()
	}
	if monitor.Keyword != nil {
		state.Keyword = types.StringValue(*monitor.Keyword)
	} else {
		state.Keyword = types.StringNull()
	}
	if monitor.MonitorGroupID != nil {
		state.MonitorGroupID = types.Int64Value(int64(*monitor.MonitorGroupID))
	} else {
		state.MonitorGroupID = types.Int64Null()
	}
	if monitor.DNSHostname != nil {
		state.DNSHostname = types.StringValue(*monitor.DNSHostname)
	} else {
		state.DNSHostname = types.StringNull()
	}
	if monitor.DNSMonitoringMode != nil {
		state.DNSMonitoringMode = types.StringValue(*monitor.DNSMonitoringMode)
	} else {
		state.DNSMonitoringMode = types.StringNull()
	}
	if monitor.Status != nil {
		state.Status = types.StringValue(*monitor.Status)
	} else {
		state.Status = types.StringNull()
	}
	if monitor.CreatedAt != nil {
		state.CreatedAt = types.StringValue(*monitor.CreatedAt)
	} else {
		state.CreatedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
