package status_page_announcements

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &statusPageAnnouncementsDataSource{}

type statusPageAnnouncementsDataSource struct {
	client *apiclient.Client
}

type statusPageAnnouncementsDataSourceModel struct {
	ID                      types.String                  `tfsdk:"id"`
	StatusPageID            types.Int64                   `tfsdk:"status_page_id"`
	StatusPageAnnouncements []statusPageAnnouncementModel `tfsdk:"status_page_announcements"`
}

type statusPageAnnouncementModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Title    types.String `tfsdk:"title"`
	Type     types.String `tfsdk:"type"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	StartsAt types.String `tfsdk:"starts_at"`
	EndsAt   types.String `tfsdk:"ends_at"`
}

func NewDataSource() datasource.DataSource {
	return &statusPageAnnouncementsDataSource{}
}

func (d *statusPageAnnouncementsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page_announcements"
}

func (d *statusPageAnnouncementsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all announcements for a status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"status_page_id": schema.Int64Attribute{
				Description: "The ID of the status page.",
				Required:    true,
			},
			"status_page_announcements": schema.ListNestedAttribute{
				Description: "List of status page announcements.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the announcement.",
							Computed:    true,
						},
						"title": schema.StringAttribute{
							Description: "The title of the announcement.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the announcement.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the announcement is enabled.",
							Computed:    true,
						},
						"starts_at": schema.StringAttribute{
							Description: "When the announcement starts.",
							Computed:    true,
						},
						"ends_at": schema.StringAttribute{
							Description: "When the announcement ends.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *statusPageAnnouncementsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *statusPageAnnouncementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state statusPageAnnouncementsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(state.StatusPageID.ValueInt64())
	announcements, err := d.client.ListStatusPageAnnouncements(statusPageID)
	if err != nil {
		resp.Diagnostics.AddError("Error listing status page announcements", err.Error())
		return
	}

	state.ID = types.StringValue("placeholder")
	state.StatusPageAnnouncements = make([]statusPageAnnouncementModel, len(announcements))

	for i, a := range announcements {
		state.StatusPageAnnouncements[i] = statusPageAnnouncementModel{
			ID:      types.Int64Value(int64(a.ID)),
			Title:   types.StringValue(a.Title),
			Type:    types.StringValue(a.Type),
			Enabled: types.BoolValue(a.Enabled),
		}
		if a.StartsAt != nil {
			state.StatusPageAnnouncements[i].StartsAt = types.StringValue(*a.StartsAt)
		} else {
			state.StatusPageAnnouncements[i].StartsAt = types.StringNull()
		}
		if a.EndsAt != nil {
			state.StatusPageAnnouncements[i].EndsAt = types.StringValue(*a.EndsAt)
		} else {
			state.StatusPageAnnouncements[i].EndsAt = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
