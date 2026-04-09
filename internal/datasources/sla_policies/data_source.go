package sla_policies

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

var _ datasource.DataSource = &slaPoliciesDataSource{}

type slaPoliciesDataSource struct {
	client *apiclient.Client
}

type slaPoliciesDataSourceModel struct {
	ID          types.String     `tfsdk:"id"`
	SlaPolicies []slaPolicyModel `tfsdk:"sla_policies"`
}

type slaPolicyModel struct {
	ID           types.Int64   `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	UptimeTarget types.Float64 `tfsdk:"uptime_target"`
	WindowType   types.String  `tfsdk:"window_type"`
	Enabled      types.Bool    `tfsdk:"enabled"`
	CreatedAt    types.String  `tfsdk:"created_at"`
}

func NewDataSource() datasource.DataSource {
	return &slaPoliciesDataSource{}
}

func (d *slaPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sla_policies"
}

func (d *slaPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all SLA policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder ID for Terraform state.",
				Computed:    true,
			},
			"sla_policies": schema.ListNestedAttribute{
				Description: "List of SLA policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the SLA policy.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the SLA policy.",
							Computed:    true,
						},
						"uptime_target": schema.Float64Attribute{
							Description: "Target uptime percentage.",
							Computed:    true,
						},
						"window_type": schema.StringAttribute{
							Description: "SLA evaluation window type.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the SLA policy is active.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "When the SLA policy was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *slaPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *slaPoliciesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	policies, err := d.client.ListSlaPolicies()
	if err != nil {
		resp.Diagnostics.AddError("Error listing SLA policies", err.Error())
		return
	}

	state := slaPoliciesDataSourceModel{
		ID:          types.StringValue("placeholder"),
		SlaPolicies: make([]slaPolicyModel, len(policies)),
	}

	for i, p := range policies {
		state.SlaPolicies[i] = slaPolicyModel{
			ID:           types.Int64Value(int64(p.ID)),
			Name:         types.StringValue(p.Name),
			UptimeTarget: types.Float64Value(p.UptimeTarget),
			WindowType:   types.StringValue(p.WindowType),
			Enabled:      types.BoolValue(p.Enabled),
			CreatedAt:    types.StringValue(p.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
