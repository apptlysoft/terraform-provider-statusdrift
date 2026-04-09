package notification_channel

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var (
	_ resource.Resource                = &notificationChannelResource{}
	_ resource.ResourceWithImportState = &notificationChannelResource{}
)

type notificationChannelResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &notificationChannelResource{}
}

func (r *notificationChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (r *notificationChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift notification channel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the notification channel.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the notification channel (max 255 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of notification channel. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"email",
						"slack",
						"microsoft_teams",
						"google_chat",
						"telegram",
						"discord",
						"pagerduty",
						"opsgenie",
						"pushover",
						"victorops",
						"datadog",
						"servicenow",
						"zapier",
						"ifttt",
						"n8n",
						"pagertree",
						"mattermost",
						"generic_webhook",
						"zenduty",
						"pushbullet",
						"push_notification",
					),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the notification channel is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"config": schema.MapAttribute{
				Description: "Configuration for the notification channel. The required keys depend on the channel type. Values are treated as sensitive.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"emails": schema.ListAttribute{
				Description: "Email addresses for email-type notification channels.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *notificationChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NotificationChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configMap, diags := configMapToAPI(ctx, plan.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	mergeEmailsIntoConfig(ctx, plan.Emails, configMap)

	input := apiclient.NotificationChannelInput{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		Config:  configMap,
	}

	result, err := r.client.CreateNotificationChannel(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating notification channel",
			fmt.Sprintf("Could not create notification channel: %s", err),
		)
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.Name = types.StringValue(result.Name)
	plan.Type = types.StringValue(result.Type)
	plan.Enabled = types.BoolValue(result.Enabled)
	// Keep the plan config as-is — the API may return masked secrets in the
	// response and we want to preserve the values the user actually provided.

	tflog.Trace(ctx, "created notification channel", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *notificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	result, err := r.client.GetNotificationChannel(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "notification channel not found, removing from state", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading notification channel",
			fmt.Sprintf("Could not read notification channel %d: %s", id, err),
		)
		return
	}

	state.ID = types.Int64Value(int64(result.ID))
	state.Name = types.StringValue(result.Name)
	state.Type = types.StringValue(result.Type)
	state.Enabled = types.BoolValue(result.Enabled)

	// Extract email array from config before merging (it can't live in map(string)).
	state.Emails = extractEmailsFromConfig(result.Config, state.Emails)

	// Merge API config with existing state config: if the API returns a masked
	// value (e.g. "****"), keep the value from the current state instead. This
	// prevents Terraform from detecting spurious diffs on sensitive fields.
	mergedConfig, diags := mergeConfigFromAPI(ctx, result.Config, state.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Config = mergedConfig

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *notificationChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NotificationChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NotificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	configMap, diags := configMapToAPI(ctx, plan.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	mergeEmailsIntoConfig(ctx, plan.Emails, configMap)

	input := apiclient.NotificationChannelInput{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		Config:  configMap,
	}

	result, err := r.client.UpdateNotificationChannel(id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating notification channel",
			fmt.Sprintf("Could not update notification channel %d: %s", id, err),
		)
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.Name = types.StringValue(result.Name)
	plan.Type = types.StringValue(result.Type)
	plan.Enabled = types.BoolValue(result.Enabled)
	// Keep the plan config — same rationale as Create.

	tflog.Trace(ctx, "updated notification channel", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *notificationChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	err := r.client.DeleteNotificationChannel(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "notification channel already deleted", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting notification channel",
			fmt.Sprintf("Could not delete notification channel %d: %s", id, err),
		)
		return
	}

	tflog.Trace(ctx, "deleted notification channel", map[string]any{"id": id})
}

func (r *notificationChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Could not parse notification channel ID %q: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// configMapToAPI converts a types.Map of String values to a map[string]interface{}
// suitable for the API client.
func configMapToAPI(ctx context.Context, tfMap types.Map) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if tfMap.IsNull() || tfMap.IsUnknown() {
		return make(map[string]interface{}), diags
	}

	elements := tfMap.Elements()
	result := make(map[string]interface{}, len(elements))
	for k, v := range elements {
		sv, ok := v.(types.String)
		if !ok {
			diags.AddError(
				"Invalid config value",
				fmt.Sprintf("Expected string value for config key %q, got %T", k, v),
			)
			continue
		}
		result[snakeToCamel(k)] = sv.ValueString()
	}

	return result, diags
}

// mergeEmailsIntoConfig adds the emails list attribute into the API config map
// as "email": ["addr1", "addr2"].
func mergeEmailsIntoConfig(ctx context.Context, emails types.List, config map[string]interface{}) {
	if emails.IsNull() || emails.IsUnknown() {
		return
	}
	var list []string
	if d := emails.ElementsAs(ctx, &list, false); !d.HasError() && len(list) > 0 {
		config["email"] = list
	}
}

// extractEmailsFromConfig pulls the "email" array out of the API response config
// and returns it as a types.List. The "email" key is deleted from apiConfig so it
// won't be processed by mergeConfigFromAPI (which expects string values).
// If "email" is absent or not an array, the existing state value is preserved.
func extractEmailsFromConfig(apiConfig map[string]interface{}, stateEmails types.List) types.List {
	raw, ok := apiConfig["email"]
	if !ok {
		return stateEmails
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return stateEmails
	}
	delete(apiConfig, "email")

	elems := make([]attr.Value, len(arr))
	for i, v := range arr {
		elems[i] = types.StringValue(fmt.Sprintf("%v", v))
	}
	listVal, d := types.ListValue(types.StringType, elems)
	if d.HasError() {
		return stateEmails
	}
	return listVal
}

// snakeToCamel converts a snake_case string to camelCase.
// e.g. "webhook_url" -> "webhookUrl"
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// camelToSnake converts a camelCase string to snake_case.
// e.g. "webhookUrl" -> "webhook_url"
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + ('a' - 'A'))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// isMaskedValue returns true if the value looks like a masked secret returned
// by the API (e.g. "****", "••••", or similar patterns).
func isMaskedValue(v string) bool {
	if len(v) == 0 {
		return false
	}
	// Fully masked: all asterisks/bullets/dots
	trimmed := strings.TrimLeft(v, "*•·")
	if len(trimmed) == 0 {
		return true
	}
	// Partially masked: first4****last4 pattern from API sanitizer
	return strings.Contains(v, "****")
}

// mergeConfigFromAPI builds a types.Map from the API response config, but
// preserves existing state values when the API value appears to be masked.
func mergeConfigFromAPI(ctx context.Context, apiConfig map[string]interface{}, stateConfig types.Map) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiConfig) == 0 {
		if stateConfig.IsNull() {
			return types.MapNull(types.StringType), diags
		}
		return stateConfig, diags
	}

	// Extract existing state values for comparison.
	stateElements := make(map[string]string)
	if !stateConfig.IsNull() && !stateConfig.IsUnknown() {
		for k, v := range stateConfig.Elements() {
			if sv, ok := v.(types.String); ok {
				stateElements[k] = sv.ValueString()
			}
		}
	}

	merged := make(map[string]attr.Value, len(apiConfig))
	for k, v := range apiConfig {
		apiVal := fmt.Sprintf("%v", v)
		stateKey := camelToSnake(k)

		// If the API returned a masked value and we have a state value for
		// this key, keep the state value to avoid spurious diffs.
		if isMaskedValue(apiVal) {
			if stateVal, exists := stateElements[stateKey]; exists {
				merged[stateKey] = types.StringValue(stateVal)
				continue
			}
		}

		merged[stateKey] = types.StringValue(apiVal)
	}

	result, d := types.MapValue(types.StringType, merged)
	diags.Append(d...)
	return result, diags
}
