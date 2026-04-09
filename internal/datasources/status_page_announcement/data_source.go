package status_page_announcement

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

var _ datasource.DataSource = &statusPageAnnouncementDataSource{}

type statusPageAnnouncementDataSource struct {
	client *apiclient.Client
}

type statusPageAnnouncementDataSourceModel struct {
	StatusPageID types.Int64  `tfsdk:"status_page_id"`
	ID           types.Int64  `tfsdk:"id"`
	Title        types.String `tfsdk:"title"`
	Content      types.String `tfsdk:"content"`
	Type         types.String `tfsdk:"type"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	StartsAt     types.String `tfsdk:"starts_at"`
	EndsAt       types.String `tfsdk:"ends_at"`
}

func NewDataSource() datasource.DataSource {
	return &statusPageAnnouncementDataSource{}
}

func (d *statusPageAnnouncementDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page_announcement"
}

func (d *statusPageAnnouncementDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single status page announcement by ID or title.",
		Attributes: map[string]schema.Attribute{
			"status_page_id": schema.Int64Attribute{
				Description: "The ID of the parent status page.",
				Required:    true,
			},
			"id": schema.Int64Attribute{
				Description: "The ID of the announcement.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("title")),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title of the announcement.",
				Optional:    true,
				Computed:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content of the announcement.",
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
	}
}

func (d *statusPageAnnouncementDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *statusPageAnnouncementDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state statusPageAnnouncementDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(state.StatusPageID.ValueInt64())
	var announcement *apiclient.StatusPageAnnouncement

	if !state.Title.IsNull() {
		title := state.Title.ValueString()
		announcements, err := d.client.ListStatusPageAnnouncements(statusPageID)
		if err != nil {
			resp.Diagnostics.AddError("Error listing status page announcements", err.Error())
			return
		}
		var matches []apiclient.StatusPageAnnouncement
		for _, a := range announcements {
			if a.Title == title {
				matches = append(matches, a)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No status page announcement found", fmt.Sprintf("No status page announcement found with title %q.", title))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple status page announcements found", fmt.Sprintf("Multiple status page announcements found with title %q, use id instead.", title))
			return
		}
		announcement = &matches[0]
	} else {
		result, err := d.client.GetStatusPageAnnouncement(statusPageID, int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading status page announcement", err.Error())
			return
		}
		announcement = result
	}

	state.ID = types.Int64Value(int64(announcement.ID))
	state.Title = types.StringValue(announcement.Title)
	state.Content = types.StringValue(announcement.Content)
	state.Type = types.StringValue(announcement.Type)
	state.Enabled = types.BoolValue(announcement.Enabled)

	if announcement.StartsAt != nil {
		state.StartsAt = types.StringValue(*announcement.StartsAt)
	} else {
		state.StartsAt = types.StringNull()
	}
	if announcement.EndsAt != nil {
		state.EndsAt = types.StringValue(*announcement.EndsAt)
	} else {
		state.EndsAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
