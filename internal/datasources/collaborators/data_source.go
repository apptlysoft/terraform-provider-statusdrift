package collaborators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &collaboratorsDataSource{}

type collaboratorsDataSource struct {
	client *apiclient.Client
}

type collaboratorsDataSourceModel struct {
	ID            types.String        `tfsdk:"id"`
	Collaborators []collaboratorModel `tfsdk:"collaborators"`
}

type collaboratorModel struct {
	ID     types.Int64  `tfsdk:"id"`
	Email  types.String `tfsdk:"email"`
	Name   types.String `tfsdk:"name"`
	Role   types.String `tfsdk:"role"`
	Status types.String `tfsdk:"status"`
}

func NewDataSource() datasource.DataSource {
	return &collaboratorsDataSource{}
}

func (d *collaboratorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collaborators"
}

func (d *collaboratorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all organization members.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"collaborators": schema.ListNestedAttribute{
				Description: "List of organization members.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the organization member.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email address.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The full name.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The organization role.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The membership status.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *collaboratorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *apiclient.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}

func (d *collaboratorsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	collabs, err := d.client.ListCollaborators()
	if err != nil {
		resp.Diagnostics.AddError("Error listing collaborators", err.Error())
		return
	}

	state := collaboratorsDataSourceModel{
		ID:            types.StringValue("placeholder"),
		Collaborators: make([]collaboratorModel, len(collabs)),
	}

	for i, c := range collabs {
		var name types.String
		if c.Name != nil {
			name = types.StringValue(*c.Name)
		} else {
			name = types.StringNull()
		}
		state.Collaborators[i] = collaboratorModel{
			ID:     types.Int64Value(int64(c.ID)),
			Email:  types.StringValue(c.Email),
			Name:   name,
			Role:   types.StringValue(c.Role),
			Status: types.StringValue(c.Status),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
