package status_pages

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &statusPagesDataSource{}

type statusPagesDataSource struct {
	client *apiclient.Client
}

type statusPagesDataSourceModel struct {
	ID          types.String      `tfsdk:"id"`
	StatusPages []statusPageModel `tfsdk:"status_pages"`
}

type statusPageModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Slug    types.String `tfsdk:"slug"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func NewDataSource() datasource.DataSource {
	return &statusPagesDataSource{}
}

func (d *statusPagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_pages"
}

func (d *statusPagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all status pages.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"status_pages": schema.ListNestedAttribute{
				Description: "List of status pages.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the status page.",
							Computed:    true,
						},
						"slug": schema.StringAttribute{
							Description: "The URL slug of the status page.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the status page.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the status page.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the status page is enabled.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *statusPagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *statusPagesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	pages, err := d.client.ListStatusPages()
	if err != nil {
		resp.Diagnostics.AddError("Error listing status pages", err.Error())
		return
	}

	state := statusPagesDataSourceModel{
		ID:          types.StringValue("placeholder"),
		StatusPages: make([]statusPageModel, len(pages)),
	}

	for i, p := range pages {
		state.StatusPages[i] = statusPageModel{
			ID:      types.Int64Value(int64(p.ID)),
			Slug:    types.StringValue(p.Slug),
			Name:    types.StringValue(p.Name),
			Type:    types.StringValue(p.Type),
			Enabled: types.BoolValue(p.Enabled),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
