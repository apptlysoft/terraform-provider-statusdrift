package locations

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &locationsDataSource{}

type locationsDataSource struct {
	client *apiclient.Client
}

type locationsDataSourceModel struct {
	ID        types.String    `tfsdk:"id"`
	Locations []locationModel `tfsdk:"locations"`
}

type locationModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	LocationKey types.String `tfsdk:"location_key"`
	IsGlobal    types.Bool   `tfsdk:"is_global"`
	Region      types.String `tfsdk:"region"`
}

func NewDataSource() datasource.DataSource {
	return &locationsDataSource{}
}

func (d *locationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locations"
}

func (d *locationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all monitoring locations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"locations": schema.ListNestedAttribute{
				Description: "List of monitoring locations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the location.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the location.",
							Computed:    true,
						},
						"location_key": schema.StringAttribute{
							Description: "The location key identifier.",
							Computed:    true,
						},
						"is_global": schema.BoolAttribute{
							Description: "Whether this is a global location.",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Description: "The region of the location.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *locationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *locationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	locations, err := d.client.ListLocations()
	if err != nil {
		resp.Diagnostics.AddError("Error listing locations", err.Error())
		return
	}

	state := locationsDataSourceModel{
		ID:        types.StringValue("placeholder"),
		Locations: make([]locationModel, len(locations)),
	}

	for i, l := range locations {
		state.Locations[i] = locationModel{
			ID:          types.Int64Value(int64(l.ID)),
			Name:        types.StringValue(l.Name),
			LocationKey: types.StringValue(l.LocationKey),
			IsGlobal:    types.BoolValue(l.IsGlobal),
		}
		if l.Region != nil {
			state.Locations[i].Region = types.StringValue(*l.Region)
		} else {
			state.Locations[i].Region = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
