package collaborator

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &collaboratorDataSource{}

type collaboratorDataSource struct {
	client *apiclient.Client
}

type collaboratorDataSourceModel struct {
	ID     types.Int64  `tfsdk:"id"`
	Email  types.String `tfsdk:"email"`
	Name   types.String `tfsdk:"name"`
	Role   types.String `tfsdk:"role"`
	Status types.String `tfsdk:"status"`
}

func NewDataSource() datasource.DataSource {
	return &collaboratorDataSource{}
}

func (d *collaboratorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_collaborator"
}

func (d *collaboratorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single organization member by ID or email. Use this to look up member IDs for on-call schedules and policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the organization member.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("email")),
				},
			},
			"email": schema.StringAttribute{
				Description: "The email address of the organization member.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The full name of the organization member.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The organization role (owner, admin, global_view, global_editor, global_communication, member).",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The membership status (pending, accepted).",
				Computed:    true,
			},
		},
	}
}

func (d *collaboratorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *collaboratorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state collaboratorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var collab *apiclient.Collaborator

	if !state.Email.IsNull() {
		email := state.Email.ValueString()
		results, err := d.client.ListCollaboratorsWithSearch(email)
		if err != nil {
			resp.Diagnostics.AddError("Error searching collaborators", err.Error())
			return
		}
		// Exact case-insensitive match on email
		var matches []apiclient.Collaborator
		for _, c := range results {
			if strings.EqualFold(c.Email, email) {
				matches = append(matches, c)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("No collaborator found", fmt.Sprintf("No collaborator found with email %q.", email))
			return
		}
		if len(matches) > 1 {
			resp.Diagnostics.AddError("Multiple collaborators found", fmt.Sprintf("Multiple collaborators found with email %q, use id instead.", email))
			return
		}
		collab = &matches[0]
	} else {
		result, err := d.client.GetCollaborator(int(state.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Error reading collaborator", err.Error())
			return
		}
		collab = result
	}

	state.ID = types.Int64Value(int64(collab.ID))
	state.Email = types.StringValue(collab.Email)
	if collab.Name != nil {
		state.Name = types.StringValue(*collab.Name)
	} else {
		state.Name = types.StringNull()
	}
	state.Role = types.StringValue(collab.Role)
	state.Status = types.StringValue(collab.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
