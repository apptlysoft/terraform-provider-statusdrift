package tags

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &tagsDataSource{}

type tagsDataSource struct {
	client *apiclient.Client
}

type tagsDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Tags []tagModel   `tfsdk:"tags"`
}

type tagModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Color types.String `tfsdk:"color"`
}

func NewDataSource() datasource.DataSource {
	return &tagsDataSource{}
}

func (d *tagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (d *tagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all tags.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"tags": schema.ListNestedAttribute{
				Description: "List of tags.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the tag.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the tag.",
							Computed:    true,
						},
						"color": schema.StringAttribute{
							Description: "The color of the tag (hex code).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *tagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tagsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tags, err := d.client.ListTags()
	if err != nil {
		resp.Diagnostics.AddError("Error listing tags", err.Error())
		return
	}

	state := tagsDataSourceModel{
		ID:   types.StringValue("placeholder"),
		Tags: make([]tagModel, len(tags)),
	}

	for i, t := range tags {
		state.Tags[i] = tagModel{
			ID:   types.Int64Value(int64(t.ID)),
			Name: types.StringValue(t.Name),
		}
		if t.Color != nil {
			state.Tags[i].Color = types.StringValue(*t.Color)
		} else {
			state.Tags[i].Color = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
