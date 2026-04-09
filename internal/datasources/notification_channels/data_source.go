package notification_channels

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &notificationChannelsDataSource{}

type notificationChannelsDataSource struct {
	client *apiclient.Client
}

type notificationChannelsDataSourceModel struct {
	ID                   types.String               `tfsdk:"id"`
	NotificationChannels []notificationChannelModel `tfsdk:"notification_channels"`
}

type notificationChannelModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func NewDataSource() datasource.DataSource {
	return &notificationChannelsDataSource{}
}

func (d *notificationChannelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channels"
}

func (d *notificationChannelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all notification channels.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"notification_channels": schema.ListNestedAttribute{
				Description: "List of notification channels.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the notification channel.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the notification channel.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the notification channel.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the notification channel is enabled.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *notificationChannelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *notificationChannelsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	channels, err := d.client.ListNotificationChannels()
	if err != nil {
		resp.Diagnostics.AddError("Error listing notification channels", err.Error())
		return
	}

	state := notificationChannelsDataSourceModel{
		ID:                   types.StringValue("placeholder"),
		NotificationChannels: make([]notificationChannelModel, len(channels)),
	}

	for i, ch := range channels {
		state.NotificationChannels[i] = notificationChannelModel{
			ID:      types.Int64Value(int64(ch.ID)),
			Name:    types.StringValue(ch.Name),
			Type:    types.StringValue(ch.Type),
			Enabled: types.BoolValue(ch.Enabled),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
