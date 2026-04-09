package api_keys

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &apiKeysDataSource{}

type apiKeysDataSource struct {
	client *apiclient.Client
}

type apiKeysDataSourceModel struct {
	ID      types.String  `tfsdk:"id"`
	APIKeys []apiKeyModel `tfsdk:"api_keys"`
}

type apiKeyModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	KeyPrefix types.String `tfsdk:"key_prefix"`
	ExpiresAt types.String `tfsdk:"expires_at"`
	IsExpired types.Bool   `tfsdk:"is_expired"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewDataSource() datasource.DataSource {
	return &apiKeysDataSource{}
}

func (d *apiKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

func (d *apiKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all API keys.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"api_keys": schema.ListNestedAttribute{
				Description: "List of API keys.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the API key.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the API key.",
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
						"is_expired": schema.BoolAttribute{
							Description: "Whether the API key has expired.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The date/time the API key was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *apiKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *apiKeysDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	keys, err := d.client.ListAPIKeys()
	if err != nil {
		resp.Diagnostics.AddError("Error listing API keys", err.Error())
		return
	}

	state := apiKeysDataSourceModel{
		ID:      types.StringValue("placeholder"),
		APIKeys: make([]apiKeyModel, len(keys)),
	}

	for i, k := range keys {
		state.APIKeys[i] = apiKeyModel{
			ID:        types.Int64Value(int64(k.ID)),
			Name:      types.StringValue(k.Name),
			KeyPrefix: types.StringValue(k.KeyPrefix),
			IsExpired: types.BoolValue(k.IsExpired),
			CreatedAt: types.StringValue(k.CreatedAt),
		}
		if k.ExpiresAt != nil {
			state.APIKeys[i].ExpiresAt = types.StringValue(*k.ExpiresAt)
		} else {
			state.APIKeys[i].ExpiresAt = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
