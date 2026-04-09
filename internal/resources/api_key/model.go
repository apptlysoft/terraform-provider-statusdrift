package api_key

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/apiclient"
)

type APIKeyModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Key           types.String `tfsdk:"key"`
	KeyPrefix     types.String `tfsdk:"key_prefix"`
	Permissions   types.Object `tfsdk:"permissions"`
	ExpiresInDays types.Int64  `tfsdk:"expires_in_days"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
	LastUsedAt    types.String `tfsdk:"last_used_at"`
	CreatedAt     types.String `tfsdk:"created_at"`
	IsExpired     types.Bool   `tfsdk:"is_expired"`
}

type PermissionsModel struct {
	Monitor                types.Object `tfsdk:"monitor"`
	MonitorGroup           types.Object `tfsdk:"monitor_group"`
	Tag                    types.Object `tfsdk:"tag"`
	NotificationChannel    types.Object `tfsdk:"notification_channel"`
	MaintenanceWindow      types.Object `tfsdk:"maintenance_window"`
	StatusPage             types.Object `tfsdk:"status_page"`
	StatusPageAnnouncement types.Object `tfsdk:"status_page_announcement"`
	Incident               types.Object `tfsdk:"incident"`
	IncidentMessage        types.Object `tfsdk:"incident_message"`
	AgentNode              types.Object `tfsdk:"agent_node"`
	ScheduledReport        types.Object `tfsdk:"scheduled_report"`
	ManualReport           types.Object `tfsdk:"manual_report"`
	SlaPolicy              types.Object `tfsdk:"sla_policy"`
	APIKey                 types.Object `tfsdk:"api_key"`
}

type ResourcePermissionsModel struct {
	Create types.Bool `tfsdk:"create"`
	Read   types.Bool `tfsdk:"read"`
	Update types.Bool `tfsdk:"update"`
	Delete types.Bool `tfsdk:"delete"`
}

// resourcePermToAPI converts a Terraform types.Object to an apiclient.ResourcePermissions.
func resourcePermToAPI(ctx context.Context, obj types.Object) (*apiclient.ResourcePermissions, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return &apiclient.ResourcePermissions{}, nil
	}

	var model ResourcePermissionsModel
	diags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &apiclient.ResourcePermissions{
		Create: model.Create.ValueBool(),
		Read:   model.Read.ValueBool(),
		Update: model.Update.ValueBool(),
		Delete: model.Delete.ValueBool(),
	}, nil
}

// permissionsToAPI converts the Terraform permissions object to the API struct.
func permissionsToAPI(ctx context.Context, obj types.Object) (*apiclient.APIKeyPermissions, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return &apiclient.APIKeyPermissions{}, nil
	}

	var model PermissionsModel
	diags := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	perms := &apiclient.APIKeyPermissions{}

	var d diag.Diagnostics

	perms.Monitor, d = resourcePermToAPI(ctx, model.Monitor)
	diags.Append(d...)
	perms.MonitorGroup, d = resourcePermToAPI(ctx, model.MonitorGroup)
	diags.Append(d...)
	perms.Tag, d = resourcePermToAPI(ctx, model.Tag)
	diags.Append(d...)
	perms.NotificationChannel, d = resourcePermToAPI(ctx, model.NotificationChannel)
	diags.Append(d...)
	perms.MaintenanceWindow, d = resourcePermToAPI(ctx, model.MaintenanceWindow)
	diags.Append(d...)
	perms.StatusPage, d = resourcePermToAPI(ctx, model.StatusPage)
	diags.Append(d...)
	perms.StatusPageAnnouncement, d = resourcePermToAPI(ctx, model.StatusPageAnnouncement)
	diags.Append(d...)
	perms.Incident, d = resourcePermToAPI(ctx, model.Incident)
	diags.Append(d...)
	perms.IncidentMessage, d = resourcePermToAPI(ctx, model.IncidentMessage)
	diags.Append(d...)
	perms.AgentNode, d = resourcePermToAPI(ctx, model.AgentNode)
	diags.Append(d...)
	perms.ScheduledReport, d = resourcePermToAPI(ctx, model.ScheduledReport)
	diags.Append(d...)
	perms.ManualReport, d = resourcePermToAPI(ctx, model.ManualReport)
	diags.Append(d...)
	perms.SlaPolicy, d = resourcePermToAPI(ctx, model.SlaPolicy)
	diags.Append(d...)
	perms.APIKey, d = resourcePermToAPI(ctx, model.APIKey)
	diags.Append(d...)

	return perms, diags
}
