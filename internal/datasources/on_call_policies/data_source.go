package on_call_policies

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &onCallPoliciesDataSource{}

type onCallPoliciesDataSource struct {
	client *apiclient.Client
}

type onCallPoliciesDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Policies []onCallPolicyModel `tfsdk:"on_call_policies"`
}

type onCallPolicyModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	IsDefault types.Bool   `tfsdk:"is_default"`
	ScopeType types.String `tfsdk:"scope_type"`
}

func NewDataSource() datasource.DataSource {
	return &onCallPoliciesDataSource{}
}

func (d *onCallPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_policies"
}

func (d *onCallPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all on-call policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"on_call_policies": schema.ListNestedAttribute{
				Description: "List of on-call policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the on-call policy.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the on-call policy.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the policy is active.",
							Computed:    true,
						},
						"is_default": schema.BoolAttribute{
							Description: "Whether this is the default policy.",
							Computed:    true,
						},
						"scope_type": schema.StringAttribute{
							Description: "Policy scope type.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *onCallPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *onCallPoliciesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	policies, err := d.client.ListOnCallPolicies()
	if err != nil {
		resp.Diagnostics.AddError("Error listing on-call policies", err.Error())
		return
	}

	state := onCallPoliciesDataSourceModel{
		ID:       types.StringValue("placeholder"),
		Policies: make([]onCallPolicyModel, len(policies)),
	}

	for i, p := range policies {
		state.Policies[i] = onCallPolicyModel{
			ID:        types.Int64Value(int64(p.ID)),
			Name:      types.StringValue(p.Name),
			Enabled:   types.BoolValue(p.Enabled),
			IsDefault: types.BoolValue(p.IsDefault),
			ScopeType: types.StringValue(p.ScopeType),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
