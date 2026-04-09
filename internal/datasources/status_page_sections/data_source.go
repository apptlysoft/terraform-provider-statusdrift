package status_page_sections

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &statusPageSectionsDataSource{}

type statusPageSectionsDataSource struct {
	client *apiclient.Client
}

type statusPageSectionsDataSourceModel struct {
	ID                 types.String             `tfsdk:"id"`
	StatusPageID       types.Int64              `tfsdk:"status_page_id"`
	StatusPageSections []statusPageSectionModel `tfsdk:"status_page_sections"`
}

type statusPageSectionModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Position types.Int64  `tfsdk:"position"`
	Expanded types.Bool   `tfsdk:"expanded"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

func NewDataSource() datasource.DataSource {
	return &statusPageSectionsDataSource{}
}

func (d *statusPageSectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page_sections"
}

func (d *statusPageSectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all sections for a status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"status_page_id": schema.Int64Attribute{
				Description: "The ID of the status page.",
				Required:    true,
			},
			"status_page_sections": schema.ListNestedAttribute{
				Description: "List of status page sections.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the section.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the section.",
							Computed:    true,
						},
						"position": schema.Int64Attribute{
							Description: "The display position.",
							Computed:    true,
						},
						"expanded": schema.BoolAttribute{
							Description: "Whether the section is expanded by default.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the section is enabled.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *statusPageSectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *statusPageSectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state statusPageSectionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(state.StatusPageID.ValueInt64())
	sections, err := d.client.ListStatusPageSections(statusPageID)
	if err != nil {
		resp.Diagnostics.AddError("Error listing status page sections", err.Error())
		return
	}

	state.ID = types.StringValue("placeholder")
	state.StatusPageSections = make([]statusPageSectionModel, len(sections))

	for i, s := range sections {
		state.StatusPageSections[i] = statusPageSectionModel{
			ID:       types.Int64Value(int64(s.ID)),
			Name:     types.StringValue(s.Name),
			Expanded: types.BoolValue(s.Expanded),
			Enabled:  types.BoolValue(s.Enabled),
		}
		if s.Position != nil {
			state.StatusPageSections[i].Position = types.Int64Value(int64(*s.Position))
		} else {
			state.StatusPageSections[i].Position = types.Int64Null()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
