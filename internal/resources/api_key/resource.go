package api_key

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

type apiKeyResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &apiKeyResource{}
}

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func resourcePermissionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"create": schema.BoolAttribute{
			Description: "Allow creating this resource type.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"read": schema.BoolAttribute{
			Description: "Allow reading this resource type.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"update": schema.BoolAttribute{
			Description: "Allow updating this resource type.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"delete": schema.BoolAttribute{
			Description: "Allow deleting this resource type.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
	}
}

func permissionBlockAttribute(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Optional:    true,
		Computed:    true,
		Attributes:  resourcePermissionAttributes(),
	}
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the API key (max 255 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"key": schema.StringAttribute{
				Description: "The full API key value. Only available after creation; cannot be retrieved later. Will be null after import.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_prefix": schema.StringAttribute{
				Description: "The first 12 characters of the API key, used for identification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_in_days": schema.Int64Attribute{
				Description: "Number of days until the API key expires. Only used at creation time. Changing this value forces resource recreation. Omit for a non-expiring key.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"expires_at": schema.StringAttribute{
				Description: "The expiration date/time of the API key in ISO 8601 format, or null if it does not expire.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used_at": schema.StringAttribute{
				Description: "The date/time the API key was last used for authentication, in ISO 8601 format.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The date/time the API key was created, in ISO 8601 format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_expired": schema.BoolAttribute{
				Description: "Whether the API key has expired.",
				Computed:    true,
			},
			"permissions": schema.SingleNestedAttribute{
				Description: "Permissions granted to this API key. Each resource type has create, read, update, and delete booleans. Omitted resources default to all false.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"monitor":                  permissionBlockAttribute("Permissions for monitor resources."),
					"monitor_group":            permissionBlockAttribute("Permissions for monitor group resources."),
					"tag":                      permissionBlockAttribute("Permissions for tag resources."),
					"notification_channel":     permissionBlockAttribute("Permissions for notification channel resources."),
					"maintenance_window":       permissionBlockAttribute("Permissions for maintenance window resources."),
					"status_page":              permissionBlockAttribute("Permissions for status page resources."),
					"status_page_announcement": permissionBlockAttribute("Permissions for status page announcement resources."),
					"incident":                 permissionBlockAttribute("Permissions for incident resources."),
					"incident_message":         permissionBlockAttribute("Permissions for incident message resources."),
					"agent_node":               permissionBlockAttribute("Permissions for agent node resources."),
					"scheduled_report":         permissionBlockAttribute("Permissions for scheduled report resources."),
					"manual_report":            permissionBlockAttribute("Permissions for manual report resources."),
					"sla_policy":               permissionBlockAttribute("Permissions for SLA policy resources."),
					"api_key":                  permissionBlockAttribute("Permissions for API key resources."),
				},
			},
		},
	}
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan APIKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	perms, diags := permissionsToAPI(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := apiclient.APIKeyCreateInput{
		Name:        plan.Name.ValueString(),
		Permissions: *perms,
	}

	input.ExpiresInDays = helpers.IntPtr(plan.ExpiresInDays)

	result, err := r.client.CreateAPIKey(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating API key",
			fmt.Sprintf("Could not create API key: %s", err),
		)
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.Name = types.StringValue(result.Name)
	plan.Key = types.StringValue(result.Key)
	plan.KeyPrefix = types.StringValue(result.KeyPrefix)
	plan.Permissions = apiclient.PermissionsFromAPI(result.Permissions)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.IsExpired = types.BoolValue(result.IsExpired)

	if result.ExpiresAt != nil {
		plan.ExpiresAt = types.StringValue(*result.ExpiresAt)
	} else {
		plan.ExpiresAt = types.StringNull()
	}

	if result.LastUsedAt != nil {
		plan.LastUsedAt = types.StringValue(*result.LastUsedAt)
	} else {
		plan.LastUsedAt = types.StringNull()
	}

	tflog.Trace(ctx, "created API key", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state APIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	result, err := r.client.GetAPIKey(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "API key not found, removing from state", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading API key",
			fmt.Sprintf("Could not read API key %d: %s", id, err),
		)
		return
	}

	// Preserve the full key from state — API only returns keyPrefix on GET.
	// state.Key keeps its existing value via UseStateForUnknown plan modifier.
	state.ID = types.Int64Value(int64(result.ID))
	state.Name = types.StringValue(result.Name)
	state.KeyPrefix = types.StringValue(result.KeyPrefix)
	state.Permissions = apiclient.PermissionsFromAPI(result.Permissions)
	state.CreatedAt = types.StringValue(result.CreatedAt)
	state.IsExpired = types.BoolValue(result.IsExpired)

	if result.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(*result.ExpiresAt)
	} else {
		state.ExpiresAt = types.StringNull()
	}

	if result.LastUsedAt != nil {
		state.LastUsedAt = types.StringValue(*result.LastUsedAt)
	} else {
		state.LastUsedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan APIKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state APIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	perms, diags := permissionsToAPI(ctx, plan.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	input := apiclient.APIKeyUpdateInput{
		Name:        &name,
		Permissions: perms,
	}

	result, err := r.client.UpdateAPIKey(id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating API key",
			fmt.Sprintf("Could not update API key %d: %s", id, err),
		)
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.Name = types.StringValue(result.Name)
	plan.Key = state.Key // Preserve key from prior state
	plan.KeyPrefix = types.StringValue(result.KeyPrefix)
	plan.Permissions = apiclient.PermissionsFromAPI(result.Permissions)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.IsExpired = types.BoolValue(result.IsExpired)

	if result.ExpiresAt != nil {
		plan.ExpiresAt = types.StringValue(*result.ExpiresAt)
	} else {
		plan.ExpiresAt = types.StringNull()
	}

	if result.LastUsedAt != nil {
		plan.LastUsedAt = types.StringValue(*result.LastUsedAt)
	} else {
		plan.LastUsedAt = types.StringNull()
	}

	tflog.Trace(ctx, "updated API key", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	err := r.client.DeleteAPIKey(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "API key already deleted", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting API key",
			fmt.Sprintf("Could not delete API key %d: %s", id, err),
		)
		return
	}

	tflog.Trace(ctx, "deleted API key", map[string]any{"id": id})
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Could not parse API key ID %q: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
