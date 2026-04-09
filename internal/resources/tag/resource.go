package tag

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
	"github.com/apptlysoft/terraform-provider-statusdrift/internal/helpers"
)

var (
	_ resource.Resource                = &tagResource{}
	_ resource.ResourceWithImportState = &tagResource{}
)

type tagResource struct {
	client *apiclient.Client
}

func NewResource() resource.Resource {
	return &tagResource{}
}

func (r *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *tagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusDrift tag.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the tag.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the tag (max 255 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"color": schema.StringAttribute{
				Description: "The hex color of the tag (e.g. #FF5733). Must match the pattern #[0-9A-Fa-f]{6}.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`),
						"must be a valid hex color (e.g. #FF5733)",
					),
				},
			},
		},
	}
}

func (r *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := apiclient.TagInput{
		Name: plan.Name.ValueString(),
	}

	input.Color = helpers.StringPtr(plan.Color)

	result, err := r.client.CreateTag(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tag",
			fmt.Sprintf("Could not create tag: %s", err),
		)
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.Name = types.StringValue(result.Name)
	if result.Color != nil {
		plan.Color = types.StringValue(*result.Color)
	} else {
		plan.Color = types.StringNull()
	}

	tflog.Trace(ctx, "created tag", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	result, err := r.client.GetTag(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "tag not found, removing from state", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading tag",
			fmt.Sprintf("Could not read tag %d: %s", id, err),
		)
		return
	}

	state.ID = types.Int64Value(int64(result.ID))
	state.Name = types.StringValue(result.Name)
	if result.Color != nil {
		state.Color = types.StringValue(*result.Color)
	} else {
		state.Color = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state TagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	input := apiclient.TagInput{
		Name: plan.Name.ValueString(),
	}

	input.Color = helpers.StringPtr(plan.Color)

	result, err := r.client.UpdateTag(id, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating tag",
			fmt.Sprintf("Could not update tag %d: %s", id, err),
		)
		return
	}

	plan.ID = types.Int64Value(int64(result.ID))
	plan.Name = types.StringValue(result.Name)
	if result.Color != nil {
		plan.Color = types.StringValue(*result.Color)
	} else {
		plan.Color = types.StringNull()
	}

	tflog.Trace(ctx, "updated tag", map[string]any{"id": result.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())

	err := r.client.DeleteTag(id)
	if err != nil {
		if apiclient.IsNotFound(err) {
			tflog.Warn(ctx, "tag already deleted", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting tag",
			fmt.Sprintf("Could not delete tag %d: %s", id, err),
		)
		return
	}

	tflog.Trace(ctx, "deleted tag", map[string]any{"id": id})
}

func (r *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Could not parse tag ID %q: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
