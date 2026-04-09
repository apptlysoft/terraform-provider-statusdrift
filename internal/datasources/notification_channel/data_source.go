package notification_channel

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

var _ datasource.DataSource = &notificationChannelDataSource{}

type notificationChannelDataSource struct {
	client *apiclient.Client
}

type notificationChannelDataSourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func NewDataSource() datasource.DataSource {
	return &notificationChannelDataSource{}
}

func (d *notificationChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (d *notificationChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single notification channel by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the notification channel.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the notification channel.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the notification channel (email, slack, webhook, etc.).",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the notification channel is enabled.",
				Computed:    true,
			},
		},
	}
}

func (d *notificationChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *notificationChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state notificationChannelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var channel *apiclient.NotificationChannel

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		channels, err := d.client.ListNotificationChannelsWithSearch(name)
		if err != nil {
			resp.Diagnostics.AddError("Error searching notification channels", err.Error())
			return
		}
		var matches []apiclient.NotificationChannel
		for _, ch := range channels {
			if ch.Name == name {
				matches = append(matches, ch)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No notification channel found", fmt.Sprintf("No notification channel found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple notification channels found", fmt.Sprintf("Multiple notification channels found with name %q, use id instead.", name))
			return
		}
		channel = &matches[0]
	} else {
		result, err := d.client.GetNotificationChannel(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading notification channel", err.Error())
			return
		}
		channel = result
	}

	state.ID = types.Int64Value(int64(channel.ID))
	state.Name = types.StringValue(channel.Name)
	state.Type = types.StringValue(channel.Type)
	state.Enabled = types.BoolValue(channel.Enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
