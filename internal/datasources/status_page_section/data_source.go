package status_page_section

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

var _ datasource.DataSource = &statusPageSectionDataSource{}

type statusPageSectionDataSource struct {
	client *apiclient.Client
}

type statusPageSectionDataSourceModel struct {
	StatusPageID types.Int64  `tfsdk:"status_page_id"`
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Position     types.Int64  `tfsdk:"position"`
	Expanded     types.Bool   `tfsdk:"expanded"`
	Enabled      types.Bool   `tfsdk:"enabled"`
}

func NewDataSource() datasource.DataSource {
	return &statusPageSectionDataSource{}
}

func (d *statusPageSectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page_section"
}

func (d *statusPageSectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single status page section by ID or name.",
		Attributes: map[string]schema.Attribute{
			"status_page_id": schema.Int64Attribute{
				Description: "The ID of the parent status page.",
				Required:    true,
			},
			"id": schema.Int64Attribute{
				Description: "The ID of the section.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the section.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the section.",
				Computed:    true,
			},
			"position": schema.Int64Attribute{
				Description: "The display position of the section.",
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
	}
}

func (d *statusPageSectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *statusPageSectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state statusPageSectionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statusPageID := int(state.StatusPageID.ValueInt64())
	var section *apiclient.StatusPageSection

	if !state.Name.IsNull() {
		name := state.Name.ValueString()
		sections, err := d.client.ListStatusPageSections(statusPageID)
		if err != nil {
			resp.Diagnostics.AddError("Error listing status page sections", err.Error())
			return
		}
		var matches []apiclient.StatusPageSection
		for _, s := range sections {
			if s.Name == name {
				matches = append(matches, s)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No status page section found", fmt.Sprintf("No status page section found with name %q.", name))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple status page sections found", fmt.Sprintf("Multiple status page sections found with name %q, use id instead.", name))
			return
		}
		section = &matches[0]
	} else {
		result, err := d.client.GetStatusPageSection(statusPageID, int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading status page section", err.Error())
			return
		}
		section = result
	}

	state.ID = types.Int64Value(int64(section.ID))
	state.Name = types.StringValue(section.Name)
	state.Expanded = types.BoolValue(section.Expanded)
	state.Enabled = types.BoolValue(section.Enabled)

	if section.Description != nil {
		state.Description = types.StringValue(*section.Description)
	} else {
		state.Description = types.StringNull()
	}
	if section.Position != nil {
		state.Position = types.Int64Value(int64(*section.Position))
	} else {
		state.Position = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
