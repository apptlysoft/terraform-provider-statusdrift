package monitor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                   = &monitorResource{}
	_ resource.ResourceWithImportState    = &monitorResource{}
	_ resource.ResourceWithValidateConfig = &monitorResource{}
)

type monitorResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &monitorResource{}
}

func (r *monitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift monitor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the monitor.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the monitor.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of monitor. Valid values: HTTPS, HTTP_KEYWORD, HTTPS_CERT, TCP, UDP, PING, PUSH, DNS.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"HTTPS",
						"HTTP_KEYWORD",
						"HTTPS_CERT",
						"TCP",
						"UDP",
						"PING",
						"PUSH",
						"DNS",
					),
				},
			},
			"interval": schema.Int64Attribute{
				Description: "The monitoring interval in minutes. Valid values depend on monitor type: non-DNS [0,1,5,10,15,30,60,1440], DNS [60,120,180,240,360,480,720,1440].",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the monitor is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"location_type": schema.StringAttribute{
				Description: "Location selection type. Valid values: locations, region.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("locations", "region"),
				},
			},
			"location_ids": schema.ListAttribute{
				Description: "List of location IDs to check from. Use this or location_names when location_type is 'locations'.",
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"location_names": schema.ListAttribute{
				Description: "List of location keys to check from (e.g. \"us-nyc-1\", \"eu-west-1\"). Resolved to IDs via the API. Use this or location_ids when location_type is 'locations'.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"region": schema.StringAttribute{
				Description: "The region to check from. Required when location_type is 'region'.",
				Optional:    true,
			},

			// HTTP fields
			"url": schema.StringAttribute{
				Description: "The URL to monitor. Required for HTTPS, HTTP_KEYWORD, and HTTPS_CERT monitor types.",
				Optional:    true,
			},
			"http_method": schema.StringAttribute{
				Description: "The HTTP method to use. Valid values: HEAD, GET, POST, PUT, PATCH, DELETE.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"),
				},
			},
			"expected_status_code": schema.StringAttribute{
				Description: "The expected HTTP status code (e.g., '200' or '200-299').",
				Optional:    true,
			},
			"http_post_body": schema.StringAttribute{
				Description: "The body to send with POST/PUT/PATCH requests.",
				Optional:    true,
			},
			"http_content_type": schema.StringAttribute{
				Description: "The Content-Type header for POST/PUT/PATCH requests.",
				Optional:    true,
			},
			"max_redirects": schema.Int64Attribute{
				Description: "Maximum number of redirects to follow (0-20).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 20),
				},
			},
			"allow_invalid_certificate": schema.BoolAttribute{
				Description: "Whether to allow invalid SSL certificates.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"cache_buster": schema.BoolAttribute{
				Description: "Whether to append a cache-busting query parameter.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"http_auth": schema.BoolAttribute{
				Description: "Whether HTTP authentication is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"http_username": schema.StringAttribute{
				Description: "The HTTP authentication username.",
				Optional:    true,
				Sensitive:   true,
			},
			"http_password": schema.StringAttribute{
				Description: "The HTTP authentication password.",
				Optional:    true,
				Sensitive:   true,
			},
			"rw_timeout": schema.Int64Attribute{
				Description: "Read/write timeout in seconds (1-45).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.Between(1, 45),
				},
			},
			"connection_timeout": schema.Int64Attribute{
				Description: "Connection timeout in seconds (1-45).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.Between(1, 45),
				},
			},

			// HTTP_KEYWORD fields
			"keyword": schema.StringAttribute{
				Description: "The keyword to search for in the response body. Used with HTTP_KEYWORD type.",
				Optional:    true,
			},
			"keyword_negation": schema.BoolAttribute{
				Description: "Whether to negate the keyword check (alert when keyword IS found).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			// TCP/UDP/PING fields
			"host": schema.StringAttribute{
				Description: "The hostname or IP to monitor. Required for TCP, UDP, and PING monitor types.",
				Optional:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port to monitor (10-65535). Required for TCP and UDP monitor types.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(10, 65535),
				},
			},

			// DNS fields
			"dns_hostname": schema.StringAttribute{
				Description: "The hostname to perform DNS lookup on. Required for DNS monitor type.",
				Optional:    true,
			},
			"dns_monitoring_mode": schema.StringAttribute{
				Description: "DNS monitoring mode. Valid values: discovery, manual.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("discovery", "manual"),
				},
			},
			"dns_servers": schema.ListAttribute{
				Description: "List of DNS servers to query (max 10).",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
			},

			// Group / tags
			"monitor_group_id": schema.Int64Attribute{
				Description: "The monitor group ID this monitor belongs to.",
				Optional:    true,
			},
			"tag_ids": schema.ListAttribute{
				Description: "List of tag IDs assigned to this monitor.",
				Optional:    true,
				ElementType: types.Int64Type,
			},

			// Alert thresholds
			"locations_down_to_alert": schema.Int64Attribute{
				Description: "Number of locations that must be down before alerting (1-10).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"checks_down_to_alert": schema.Int64Attribute{
				Description: "Number of consecutive checks that must fail before alerting (1-10).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"warning_threshold_ms": schema.Int64Attribute{
				Description: "Response time threshold in milliseconds for warning status (100-60000).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20000),
				Validators: []validator.Int64{
					int64validator.Between(100, 60000),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"alert": schema.ListNestedBlock{
				Description: "Alert configuration blocks.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"notification_channel_ids": schema.ListAttribute{
							Description: "List of notification channel IDs to notify.",
							Required:    true,
							ElementType: types.Int64Type,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
						},
						"conditions": schema.ListAttribute{
							Description: "List of alert conditions.",
							Required:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
						},
						"delay_minutes": schema.Int64Attribute{
							Description: "Delay in minutes before sending the alert.",
							Optional:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether this alert is enabled.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
						},
						"recurring_interval_minutes": schema.Int64Attribute{
							Description: "Interval in minutes for recurring alerts.",
							Optional:    true,
						},
						"max_recurrences": schema.Int64Attribute{
							Description: "Maximum number of recurring alert notifications.",
							Optional:    true,
						},
						"notify_on_call": schema.BoolAttribute{
							Description: "Whether this alert should trigger the on-call escalation policy.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"on_call_policy_id": schema.Int64Attribute{
							Description: "On-call policy ID to use for this alert. When null and notify_on_call is true, the organization default policy is used.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

// ValidateConfig validates the monitor configuration based on its type.
func (r *monitorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data MonitorModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If type is unknown (e.g. a variable), we cannot validate type-specific fields.
	if data.Type.IsUnknown() {
		return
	}

	monitorType := data.Type.ValueString()

	switch monitorType {
	case "HTTPS", "HTTP_KEYWORD", "HTTPS_CERT":
		if data.URL.IsNull() || data.URL.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("url"),
				"Missing Required Attribute",
				fmt.Sprintf("The 'url' attribute is required for monitor type %q.", monitorType),
			)
		}
	case "TCP", "UDP":
		if data.Host.IsNull() || data.Host.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("host"),
				"Missing Required Attribute",
				fmt.Sprintf("The 'host' attribute is required for monitor type %q.", monitorType),
			)
		}
		if data.Port.IsNull() || data.Port.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("port"),
				"Missing Required Attribute",
				fmt.Sprintf("The 'port' attribute is required for monitor type %q.", monitorType),
			)
		}
	case "PING":
		if data.Host.IsNull() || data.Host.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("host"),
				"Missing Required Attribute",
				"The 'host' attribute is required for monitor type \"PING\".",
			)
		}
	case "DNS":
		if data.DNSHostname.IsNull() || data.DNSHostname.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("dns_hostname"),
				"Missing Required Attribute",
				"The 'dns_hostname' attribute is required for monitor type \"DNS\".",
			)
		}
	}

	// Validate interval values based on type.
	if !data.Interval.IsNull() && !data.Interval.IsUnknown() {
		interval := data.Interval.ValueInt64()
		if monitorType == "DNS" {
			validDNSIntervals := map[int64]bool{60: true, 120: true, 180: true, 240: true, 360: true, 480: true, 720: true, 1440: true}
			if !validDNSIntervals[interval] {
				resp.Diagnostics.AddAttributeError(
					path.Root("interval"),
					"Invalid Interval for DNS Monitor",
					"DNS monitors support intervals: 60, 120, 180, 240, 360, 480, 720, 1440.",
				)
			}
		} else {
			validIntervals := map[int64]bool{0: true, 1: true, 5: true, 10: true, 15: true, 30: true, 60: true, 1440: true}
			if !validIntervals[interval] {
				resp.Diagnostics.AddAttributeError(
					path.Root("interval"),
					"Invalid Interval",
					"Non-DNS monitors support intervals: 0, 1, 5, 10, 15, 30, 60, 1440.",
				)
			}
		}
	}

	// Validate location_ids and location_names are mutually exclusive
	idsSet := !data.LocationIDs.IsNull()
	namesSet := !data.LocationNames.IsNull()
	if idsSet && namesSet {
		resp.Diagnostics.AddError(
			"Conflicting Location Configuration",
			"Only one of 'location_ids' or 'location_names' may be specified, not both.",
		)
	}

	// When location_type is "locations", one of location_ids or location_names is required
	// Skip if either value is unknown (e.g. from a variable) — will be validated at plan/apply time
	idsUnknown := data.LocationIDs.IsUnknown()
	namesUnknown := data.LocationNames.IsUnknown()
	if !data.LocationType.IsUnknown() && data.LocationType.ValueString() == "locations" && !idsSet && !namesSet && !idsUnknown && !namesUnknown {
		resp.Diagnostics.AddError(
			"Missing Location Configuration",
			"One of 'location_ids' or 'location_names' is required when location_type is 'locations'.",
		)
	}
}

func (r *monitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new monitor.
func (r *monitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := modelToInput(ctx, &plan, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating monitor", map[string]interface{}{
		"name": input.Name,
		"type": input.Type,
	})

	monitor, err := r.client.CreateMonitor(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Monitor",
			fmt.Sprintf("Could not create monitor: %s", err),
		)
		return
	}

	diags = apiToModel(ctx, monitor, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the current state of a monitor.
func (r *monitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	tflog.Debug(ctx, "Reading monitor", map[string]interface{}{
		"id": id,
	})

	monitor, err := r.client.GetMonitor(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "Monitor not found, removing from state", map[string]interface{}{
				"id": id,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Monitor",
			fmt.Sprintf("Could not read monitor %d: %s", id, err),
		)
		return
	}

	diags := apiToModel(ctx, monitor, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an existing monitor.
func (r *monitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MonitorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(plan.ID.ValueInt64())

	input, diags := modelToInput(ctx, &plan, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating monitor", map[string]interface{}{
		"id": id,
	})

	monitor, err := r.client.UpdateMonitor(id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Monitor",
			fmt.Sprintf("Could not update monitor %d: %s", id, err),
		)
		return
	}

	diags = apiToModel(ctx, monitor, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes a monitor.
func (r *monitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MonitorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	tflog.Debug(ctx, "Deleting monitor", map[string]interface{}{
		"id": id,
	})

	err := r.client.DeleteMonitor(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			// Already deleted.
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Monitor",
			fmt.Sprintf("Could not delete monitor %d: %s", id, err),
		)
	}
}

// ImportState imports a monitor by its numeric ID.
func (r *monitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("The monitor ID must be a numeric value, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// ---------------------------------------------------------------------------
// Model-to-API conversion
// ---------------------------------------------------------------------------

func modelToInput(ctx context.Context, m *MonitorModel, client *apiclient.Client) (apiclient.MonitorInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := apiclient.MonitorInput{
		Name:                 m.Name.ValueString(),
		Type:                 m.Type.ValueString(),
		Interval:             int(m.Interval.ValueInt64()),
		Enabled:              m.Enabled.ValueBool(),
		LocationType:         m.LocationType.ValueString(),
		AllowInvalidCert:     m.AllowInvalidCert.ValueBool(),
		CacheBuster:          m.CacheBuster.ValueBool(),
		HTTPAuth:             m.HTTPAuth.ValueBool(),
		KeywordNegation:      m.KeywordNegation.ValueBool(),
		LocationsDownToAlert: int(m.LocationsDownToAlert.ValueInt64()),
		ChecksDownToAlert:    int(m.ChecksDownToAlert.ValueInt64()),
		WarningThresholdMs:   int(m.WarningThresholdMs.ValueInt64()),
	}

	// LocationIDs — from explicit IDs
	if !m.LocationIDs.IsNull() && !m.LocationIDs.IsUnknown() {
		var ids []int64
		diags.Append(m.LocationIDs.ElementsAs(ctx, &ids, false)...)
		intIDs := make([]int, len(ids))
		for i, v := range ids {
			intIDs[i] = int(v)
		}
		input.LocationIDs = intIDs
	}

	// LocationNames — resolve names to IDs via API
	if !m.LocationNames.IsNull() && !m.LocationNames.IsUnknown() {
		var names []string
		diags.Append(m.LocationNames.ElementsAs(ctx, &names, false)...)
		if diags.HasError() {
			return input, diags
		}

		locations, err := client.ListLocations()
		if err != nil {
			diags.AddError(
				"Error Resolving Location Names",
				fmt.Sprintf("Could not list locations: %s", err),
			)
			return input, diags
		}

		keyToID := make(map[string]int, len(locations))
		for _, loc := range locations {
			keyToID[loc.LocationKey] = loc.ID
		}

		resolvedIDs := make([]int, 0, len(names))
		for _, name := range names {
			id, ok := keyToID[name]
			if !ok {
				diags.AddAttributeError(
					path.Root("location_names"),
					"Unknown Location Name",
					fmt.Sprintf("Location key %q not found. Available locations: %s", name, availableLocationKeys(locations)),
				)
				continue
			}
			resolvedIDs = append(resolvedIDs, id)
		}
		if diags.HasError() {
			return input, diags
		}

		input.LocationIDs = resolvedIDs
	}

	// Region
	input.Region = helpers.StringPtr(m.Region)

	// URL
	input.URL = helpers.StringPtr(m.URL)

	// Host
	input.Host = helpers.StringPtr(m.Host)

	// Port
	input.Port = helpers.IntPtr(m.Port)

	// HTTPMethod
	input.HTTPMethod = helpers.StringPtr(m.HTTPMethod)

	// HTTPPostBody
	input.HTTPPostBody = helpers.StringPtr(m.HTTPPostBody)

	// HTTPContentType
	input.HTTPContentType = helpers.StringPtr(m.HTTPContentType)

	// MaxRedirects
	input.MaxRedirects = helpers.IntPtr(m.MaxRedirects)

	// HTTPUsername
	input.HTTPUsername = helpers.StringPtr(m.HTTPUsername)

	// HTTPPassword
	input.HTTPPassword = helpers.StringPtr(m.HTTPPassword)

	// RWTimeout
	input.RWTimeout = helpers.IntPtr(m.RWTimeout)

	// ConnectionTimeout
	input.ConnectionTimeout = helpers.IntPtr(m.ConnectionTimeout)

	// ExpectedStatusCode
	input.ExpectedStatusCode = helpers.StringPtr(m.ExpectedStatusCode)

	// Keyword
	input.Keyword = helpers.StringPtr(m.Keyword)

	// MonitorGroupID
	input.MonitorGroupID = helpers.IntPtr(m.MonitorGroupID)

	// TagIDs
	if !m.TagIDs.IsNull() && !m.TagIDs.IsUnknown() {
		var ids []int64
		diags.Append(m.TagIDs.ElementsAs(ctx, &ids, false)...)
		intIDs := make([]int, len(ids))
		for i, v := range ids {
			intIDs[i] = int(v)
		}
		input.TagIDs = intIDs
	}

	// DNSHostname
	input.DNSHostname = helpers.StringPtr(m.DNSHostname)

	// DNSMonitoringMode
	input.DNSMonitoringMode = helpers.StringPtr(m.DNSMonitoringMode)

	// DNSServers
	if !m.DNSServers.IsNull() && !m.DNSServers.IsUnknown() {
		var servers []string
		diags.Append(m.DNSServers.ElementsAs(ctx, &servers, false)...)
		input.DNSServers = servers
	}

	// Alerts
	if !m.Alerts.IsNull() && !m.Alerts.IsUnknown() {
		var alertModels []AlertModel
		diags.Append(m.Alerts.ElementsAs(ctx, &alertModels, false)...)

		alerts := make([]apiclient.AlertConfig, 0, len(alertModels))
		for _, am := range alertModels {
			ac := apiclient.AlertConfig{
				Enabled: am.Enabled.ValueBool(),
			}

			// NotificationChannelIDs
			if !am.NotificationChannelIDs.IsNull() && !am.NotificationChannelIDs.IsUnknown() {
				var ids []int64
				diags.Append(am.NotificationChannelIDs.ElementsAs(ctx, &ids, false)...)
				intIDs := make([]int, len(ids))
				for i, v := range ids {
					intIDs[i] = int(v)
				}
				ac.NotificationChannelIDs = intIDs
			}

			// Conditions
			if !am.Conditions.IsNull() && !am.Conditions.IsUnknown() {
				var conditions []string
				diags.Append(am.Conditions.ElementsAs(ctx, &conditions, false)...)
				ac.Conditions = conditions
			}

			// DelayMinutes
			ac.DelayMinutes = helpers.IntPtr(am.DelayMinutes)

			// RecurringIntervalMinutes
			ac.RecurringIntervalMinutes = helpers.IntPtr(am.RecurringIntervalMinutes)

			// MaxRecurrences
			ac.MaxRecurrences = helpers.IntPtr(am.MaxRecurrences)

			// On-Call
			ac.NotifyOnCall = am.NotifyOnCall.ValueBool()
			ac.OnCallPolicyID = helpers.IntPtr(am.OnCallPolicyID)

			alerts = append(alerts, ac)
		}
		input.Alerts = alerts
	}

	return input, diags
}

func availableLocationKeys(locations []apiclient.Location) string {
	keys := make([]string, len(locations))
	for i, loc := range locations {
		keys[i] = fmt.Sprintf("%q", loc.LocationKey)
	}
	return fmt.Sprintf("[%s]", strings.Join(keys, ", "))
}

// ---------------------------------------------------------------------------
// API-to-Model conversion
// ---------------------------------------------------------------------------

func apiToModel(ctx context.Context, monitor *apiclient.Monitor, m *MonitorModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.Int64Value(int64(monitor.ID))
	m.Name = types.StringValue(monitor.Name)
	m.Type = types.StringValue(monitor.Type)
	m.Interval = types.Int64Value(int64(monitor.Interval))
	m.Enabled = types.BoolValue(monitor.Enabled)
	m.LocationType = types.StringValue(monitor.LocationType)

	// LocationIDs — update from API when available, otherwise preserve plan/state value.
	// When location_names is used, skip entirely (IDs are resolved at apply time).
	usesNames := !m.LocationNames.IsNull() && !m.LocationNames.IsUnknown()
	if !usesNames {
		if len(monitor.LocationIDs) > 0 {
			ids := make([]attr.Value, len(monitor.LocationIDs))
			for i, v := range monitor.LocationIDs {
				ids[i] = types.Int64Value(int64(v))
			}
			listVal, d := types.ListValue(types.Int64Type, ids)
			diags.Append(d...)
			m.LocationIDs = listVal
		} else if m.LocationIDs.IsNull() || m.LocationIDs.IsUnknown() {
			m.LocationIDs = types.ListNull(types.Int64Type)
		}
		// else: API returned empty but plan/state has IDs — keep existing value
	}

	// Region
	if monitor.Region != nil {
		m.Region = types.StringValue(*monitor.Region)
	} else {
		m.Region = types.StringNull()
	}

	// URL
	if monitor.URL != nil {
		m.URL = types.StringValue(*monitor.URL)
	} else {
		m.URL = types.StringNull()
	}

	// Host
	if monitor.Host != nil {
		m.Host = types.StringValue(*monitor.Host)
	} else {
		m.Host = types.StringNull()
	}

	// Port
	if monitor.Port != nil {
		m.Port = types.Int64Value(int64(*monitor.Port))
	} else {
		m.Port = types.Int64Null()
	}

	// HTTP-specific fields: only update from API when the user configured them
	// (the API may return default values for these even on non-HTTP monitors).

	// HTTPMethod
	if !m.HTTPMethod.IsNull() && monitor.HTTPMethod != nil {
		m.HTTPMethod = types.StringValue(*monitor.HTTPMethod)
	} else if m.HTTPMethod.IsNull() {
		// keep null
	} else {
		m.HTTPMethod = types.StringNull()
	}

	// HTTPPostBody
	if !m.HTTPPostBody.IsNull() && monitor.HTTPPostBody != nil {
		m.HTTPPostBody = types.StringValue(*monitor.HTTPPostBody)
	} else if m.HTTPPostBody.IsNull() {
		// keep null
	} else {
		m.HTTPPostBody = types.StringNull()
	}

	// HTTPContentType
	if !m.HTTPContentType.IsNull() && monitor.HTTPContentType != nil {
		m.HTTPContentType = types.StringValue(*monitor.HTTPContentType)
	} else if m.HTTPContentType.IsNull() {
		// keep null
	} else {
		m.HTTPContentType = types.StringNull()
	}

	// MaxRedirects
	if !m.MaxRedirects.IsNull() && monitor.MaxRedirects != nil {
		m.MaxRedirects = types.Int64Value(int64(*monitor.MaxRedirects))
	} else if m.MaxRedirects.IsNull() {
		// keep null
	} else {
		m.MaxRedirects = types.Int64Null()
	}

	// Bool fields: only update from API when the user configured them
	if !m.AllowInvalidCert.IsNull() {
		m.AllowInvalidCert = types.BoolValue(monitor.AllowInvalidCert)
	}
	if !m.CacheBuster.IsNull() {
		m.CacheBuster = types.BoolValue(monitor.CacheBuster)
	}
	if !m.HTTPAuth.IsNull() {
		m.HTTPAuth = types.BoolValue(monitor.HTTPAuth)
	}
	if !m.KeywordNegation.IsNull() {
		m.KeywordNegation = types.BoolValue(monitor.KeywordNegation)
	}

	// HTTPUsername
	if !m.HTTPUsername.IsNull() && monitor.HTTPUsername != nil {
		m.HTTPUsername = types.StringValue(*monitor.HTTPUsername)
	} else if m.HTTPUsername.IsNull() {
		// keep null
	} else {
		m.HTTPUsername = types.StringNull()
	}

	// HTTPPassword - the API typically does not return the password.
	// Preserve the planned/state value to avoid spurious diffs.
	if monitor.HTTPPassword != nil && *monitor.HTTPPassword != "" {
		m.HTTPPassword = types.StringValue(*monitor.HTTPPassword)
	} else if m.HTTPPassword.IsNull() || m.HTTPPassword.IsUnknown() {
		m.HTTPPassword = types.StringNull()
	}
	// else: keep existing value in state

	// RWTimeout
	if monitor.RWTimeout != nil {
		m.RWTimeout = types.Int64Value(int64(*monitor.RWTimeout))
	} else {
		m.RWTimeout = types.Int64Null()
	}

	// ConnectionTimeout
	if monitor.ConnectionTimeout != nil {
		m.ConnectionTimeout = types.Int64Value(int64(*monitor.ConnectionTimeout))
	} else {
		m.ConnectionTimeout = types.Int64Null()
	}

	// ExpectedStatusCode
	if !m.ExpectedStatusCode.IsNull() && monitor.ExpectedStatusCode != nil {
		m.ExpectedStatusCode = types.StringValue(*monitor.ExpectedStatusCode)
	} else if m.ExpectedStatusCode.IsNull() {
		// keep null
	} else {
		m.ExpectedStatusCode = types.StringNull()
	}

	// Keyword
	if !m.Keyword.IsNull() && monitor.Keyword != nil {
		m.Keyword = types.StringValue(*monitor.Keyword)
	} else if m.Keyword.IsNull() {
		// keep null
	} else {
		m.Keyword = types.StringNull()
	}

	// MonitorGroupID — preserve plan/state value if API returns nil
	if monitor.MonitorGroupID != nil {
		m.MonitorGroupID = types.Int64Value(int64(*monitor.MonitorGroupID))
	} else if m.MonitorGroupID.IsNull() || m.MonitorGroupID.IsUnknown() {
		m.MonitorGroupID = types.Int64Null()
	}

	// TagIDs — preserve plan/state value if API returns empty
	if len(monitor.TagIDs) > 0 {
		ids := make([]attr.Value, len(monitor.TagIDs))
		for i, v := range monitor.TagIDs {
			ids[i] = types.Int64Value(int64(v))
		}
		listVal, d := types.ListValue(types.Int64Type, ids)
		diags.Append(d...)
		m.TagIDs = listVal
	} else if m.TagIDs.IsNull() || m.TagIDs.IsUnknown() {
		m.TagIDs = types.ListNull(types.Int64Type)
	}

	// Alert thresholds
	m.LocationsDownToAlert = types.Int64Value(int64(monitor.LocationsDownToAlert))
	m.ChecksDownToAlert = types.Int64Value(int64(monitor.ChecksDownToAlert))
	m.WarningThresholdMs = types.Int64Value(int64(monitor.WarningThresholdMs))

	// DNSHostname
	if monitor.DNSHostname != nil {
		m.DNSHostname = types.StringValue(*monitor.DNSHostname)
	} else {
		m.DNSHostname = types.StringNull()
	}

	// DNSMonitoringMode — the API may normalize synonyms (e.g. "discovered" → "discovery").
	// Preserve the existing state value when the meanings are equivalent.
	if monitor.DNSMonitoringMode != nil {
		apiVal := *monitor.DNSMonitoringMode
		if !m.DNSMonitoringMode.IsNull() && dnsMonitoringModeEquiv(m.DNSMonitoringMode.ValueString(), apiVal) {
			// keep existing state value
		} else {
			m.DNSMonitoringMode = types.StringValue(apiVal)
		}
	} else {
		m.DNSMonitoringMode = types.StringNull()
	}

	// DNSServers
	if len(monitor.DNSServers) > 0 {
		servers := make([]attr.Value, len(monitor.DNSServers))
		for i, v := range monitor.DNSServers {
			servers[i] = types.StringValue(v)
		}
		listVal, d := types.ListValue(types.StringType, servers)
		diags.Append(d...)
		m.DNSServers = listVal
	} else {
		m.DNSServers = types.ListNull(types.StringType)
	}

	// Alerts
	if len(monitor.Alerts) > 0 {
		alertAttrTypes := map[string]attr.Type{
			"notification_channel_ids":   types.ListType{ElemType: types.Int64Type},
			"conditions":                 types.ListType{ElemType: types.StringType},
			"delay_minutes":              types.Int64Type,
			"enabled":                    types.BoolType,
			"recurring_interval_minutes": types.Int64Type,
			"max_recurrences":            types.Int64Type,
			"notify_on_call":             types.BoolType,
			"on_call_policy_id":          types.Int64Type,
		}

		alertValues := make([]attr.Value, 0, len(monitor.Alerts))
		for alertIdx, ac := range monitor.Alerts {
			alertAttrs := map[string]attr.Value{}

			// NotificationChannelIDs — preserve the existing order from
			// plan/state so the API returning IDs in a different order
			// doesn't cause spurious diffs.
			if len(ac.NotificationChannelIDs) > 0 {
				apiIDs := ac.NotificationChannelIDs
				if !m.Alerts.IsNull() && !m.Alerts.IsUnknown() {
					var existingAlerts []AlertModel
					if d := m.Alerts.ElementsAs(ctx, &existingAlerts, false); !d.HasError() && alertIdx < len(existingAlerts) {
						apiIDs = matchIDOrder(apiIDs, existingAlerts[alertIdx])
					}
				}
				ids := make([]attr.Value, len(apiIDs))
				for i, v := range apiIDs {
					ids[i] = types.Int64Value(int64(v))
				}
				listVal, d := types.ListValue(types.Int64Type, ids)
				diags.Append(d...)
				alertAttrs["notification_channel_ids"] = listVal
			} else {
				alertAttrs["notification_channel_ids"] = types.ListNull(types.Int64Type)
			}

			// Conditions
			if len(ac.Conditions) > 0 {
				conds := make([]attr.Value, len(ac.Conditions))
				for i, v := range ac.Conditions {
					conds[i] = types.StringValue(v)
				}
				listVal, d := types.ListValue(types.StringType, conds)
				diags.Append(d...)
				alertAttrs["conditions"] = listVal
			} else {
				alertAttrs["conditions"] = types.ListNull(types.StringType)
			}

			// DelayMinutes
			if ac.DelayMinutes != nil {
				alertAttrs["delay_minutes"] = types.Int64Value(int64(*ac.DelayMinutes))
			} else {
				alertAttrs["delay_minutes"] = types.Int64Null()
			}

			// Enabled
			alertAttrs["enabled"] = types.BoolValue(ac.Enabled)

			// RecurringIntervalMinutes
			if ac.RecurringIntervalMinutes != nil {
				alertAttrs["recurring_interval_minutes"] = types.Int64Value(int64(*ac.RecurringIntervalMinutes))
			} else {
				alertAttrs["recurring_interval_minutes"] = types.Int64Null()
			}

			// MaxRecurrences
			if ac.MaxRecurrences != nil {
				alertAttrs["max_recurrences"] = types.Int64Value(int64(*ac.MaxRecurrences))
			} else {
				alertAttrs["max_recurrences"] = types.Int64Null()
			}

			// On-Call
			alertAttrs["notify_on_call"] = types.BoolValue(ac.NotifyOnCall)
			if ac.OnCallPolicyID != nil {
				alertAttrs["on_call_policy_id"] = types.Int64Value(int64(*ac.OnCallPolicyID))
			} else {
				alertAttrs["on_call_policy_id"] = types.Int64Null()
			}

			objVal, d := types.ObjectValue(alertAttrTypes, alertAttrs)
			diags.Append(d...)
			alertValues = append(alertValues, objVal)
		}

		alertObjType := types.ObjectType{
			AttrTypes: alertAttrTypes,
		}
		listVal, d := types.ListValue(alertObjType, alertValues)
		diags.Append(d...)
		m.Alerts = listVal
	} else {
		alertAttrTypes := map[string]attr.Type{
			"notification_channel_ids":   types.ListType{ElemType: types.Int64Type},
			"conditions":                 types.ListType{ElemType: types.StringType},
			"delay_minutes":              types.Int64Type,
			"enabled":                    types.BoolType,
			"recurring_interval_minutes": types.Int64Type,
			"max_recurrences":            types.Int64Type,
			"notify_on_call":             types.BoolType,
			"on_call_policy_id":          types.Int64Type,
		}
		alertObjType := types.ObjectType{
			AttrTypes: alertAttrTypes,
		}
		m.Alerts = types.ListNull(alertObjType)
	}

	return diags
}

// matchIDOrder reorders apiIDs to match the order of IDs in the existing alert
// model. If the sets differ (IDs added/removed), it returns apiIDs as-is.
func matchIDOrder(apiIDs []int, existing AlertModel) []int {
	if existing.NotificationChannelIDs.IsNull() || existing.NotificationChannelIDs.IsUnknown() {
		return apiIDs
	}

	var existingIDs []int64
	if d := existing.NotificationChannelIDs.ElementsAs(context.Background(), &existingIDs, false); d.HasError() {
		return apiIDs
	}

	if len(existingIDs) != len(apiIDs) {
		return apiIDs
	}

	// Build a set of API IDs for quick lookup.
	apiSet := make(map[int]bool, len(apiIDs))
	for _, id := range apiIDs {
		apiSet[id] = true
	}

	// Verify the sets are identical.
	for _, id := range existingIDs {
		if !apiSet[int(id)] {
			return apiIDs
		}
	}

	// Same set — use existing order.
	result := make([]int, len(existingIDs))
	for i, id := range existingIDs {
		result[i] = int(id)
	}
	return result
}

// dnsMonitoringModeEquiv returns true when two dns_monitoring_mode values are
// synonyms that the API treats identically (e.g. "discovered" and "discovery").
func dnsMonitoringModeEquiv(a, b string) bool {
	return normalizeDNSMode(a) == normalizeDNSMode(b)
}

func normalizeDNSMode(v string) string {
	if v == "discovered" || v == "discovery" {
		return "discovery"
	}
	return v
}
