package api_key

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

var _ datasource.DataSource = &apiKeyDataSource{}

type apiKeyDataSource struct {
	client *apiclient.Client
}

type apiKeyDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	KeyPrefix   types.String `tfsdk:"key_prefix"`
	Permissions types.Object `tfsdk:"permissions"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	LastUsedAt  types.String `tfsdk:"last_used_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	IsExpired   types.Bool   `tfsdk:"is_expired"`
}

func NewDataSource() datasource.DataSource {
	return &apiKeyDataSource{}
}

func (d *apiKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func computedPermissionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"create": schema.BoolAttribute{Description: "Allow creating this resource type.", Computed: true},
		"read":   schema.BoolAttribute{Description: "Allow reading this resource type.", Computed: true},
		"update": schema.BoolAttribute{Description: "Allow updating this resource type.", Computed: true},
		"delete": schema.BoolAttribute{Description: "Allow deleting this resource type.", Computed: true},
	}
}

func computedPermissionBlock(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Computed:    true,
		Attributes:  computedPermissionAttributes(),
	}
}

func (d *apiKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single API key by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the API key.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the API key.",
				Optional:    true,
				Computed:    true,
			},
			"key_prefix": schema.StringAttribute{
				Description: "The first 12 characters of the API key.",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "The expiration date/time in ISO 8601 format, or null if non-expiring.",
				Computed:    true,
			},
			"last_used_at": schema.StringAttribute{
				Description: "The date/time the API key was last used.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The date/time the API key was created.",
				Computed:    true,
			},
			"is_expired": schema.BoolAttribute{
				Description: "Whether the API key has expired.",
				Computed:    true,
			},
			"permissions": schema.SingleNestedAttribute{
				Description: "Permissions granted to this API key.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"monitor":                  computedPermissionBlock("Permissions for monitor resources."),
					"monitor_group":            computedPermissionBlock("Permissions for monitor group resources."),
					"tag":                      computedPermissionBlock("Permissions for tag resources."),
					"notification_channel":     computedPermissionBlock("Permissions for notification channel resources."),
					"maintenance_window":       computedPermissionBlock("Permissions for maintenance window resources."),
					"status_page":              computedPermissionBlock("Permissions for status page resources."),
					"status_page_announcement": computedPermissionBlock("Permissions for status page announcement resources."),
					"incident":                 computedPermissionBlock("Permissions for incident resources."),
					"incident_message":         computedPermissionBlock("Permissions for incident message resources."),
					"agent_node":               computedPermissionBlock("Permissions for agent node resources."),
					"scheduled_report":         computedPermissionBlock("Permissions for scheduled report resources."),
					"manual_report":            computedPermissionBlock("Permissions for manual report resources."),
					"sla_policy":               computedPermissionBlock("Permissions for SLA policy resources."),
					"api_key":                  computedPermissionBlock("Permissions for API key resources."),
				},
			},
		},
	}
}

func (d *apiKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *apiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state apiKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiKey *apiclient.APIKey

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		keys, err := d.client.ListAPIKeysWithSearch(name)
		if err != nil {
			resp.Diagnostics.AddError("Error searching API keys", err.Error())
			return
		}
		var matches []apiclient.APIKey
		for _, k := range keys {
			if k.Name == name {
				matches = append(matches, k)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No API key found", fmt.Sprintf("No API key found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple API keys found", fmt.Sprintf("Multiple API keys found with name %q, use id instead.", name))
			return
		}
		apiKey = &matches[0]
	} else {
		result, err := d.client.GetAPIKey(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading API key", err.Error())
			return
		}
		apiKey = result
	}

	state.ID = types.Int64Value(int64(apiKey.ID))
	state.Name = types.StringValue(apiKey.Name)
	state.KeyPrefix = types.StringValue(apiKey.KeyPrefix)
	state.Permissions = apiclient.PermissionsFromAPI(apiKey.Permissions)
	state.CreatedAt = types.StringValue(apiKey.CreatedAt)
	state.IsExpired = types.BoolValue(apiKey.IsExpired)

	if apiKey.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(*apiKey.ExpiresAt)
	} else {
		state.ExpiresAt = types.StringNull()
	}

	if apiKey.LastUsedAt != nil {
		state.LastUsedAt = types.StringValue(*apiKey.LastUsedAt)
	} else {
		state.LastUsedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
