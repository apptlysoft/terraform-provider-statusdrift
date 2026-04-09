package apiclient

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ResourcePermissionsAttrTypes returns the attr.Type map for a single resource's CRUD permissions.
func ResourcePermissionsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"create": types.BoolType,
		"read":   types.BoolType,
		"update": types.BoolType,
		"delete": types.BoolType,
	}
}

// PermissionsAttrTypes returns the attr.Type map for the full API key permissions object.
func PermissionsAttrTypes() map[string]attr.Type {
	rpType := types.ObjectType{AttrTypes: ResourcePermissionsAttrTypes()}
	return map[string]attr.Type{
		"monitor":                  rpType,
		"monitor_group":            rpType,
		"tag":                      rpType,
		"notification_channel":     rpType,
		"maintenance_window":       rpType,
		"status_page":              rpType,
		"status_page_announcement": rpType,
		"incident":                 rpType,
		"incident_message":         rpType,
		"agent_node":               rpType,
		"scheduled_report":         rpType,
		"manual_report":            rpType,
		"sla_policy":               rpType,
		"api_key":                  rpType,
	}
}

// ResourcePermFromAPI converts a ResourcePermissions to a Terraform types.Object.
func ResourcePermFromAPI(rp *ResourcePermissions) types.Object {
	if rp == nil {
		rp = &ResourcePermissions{}
	}

	return types.ObjectValueMust(ResourcePermissionsAttrTypes(), map[string]attr.Value{
		"create": types.BoolValue(rp.Create),
		"read":   types.BoolValue(rp.Read),
		"update": types.BoolValue(rp.Update),
		"delete": types.BoolValue(rp.Delete),
	})
}

// PermissionsFromAPI converts an APIKeyPermissions struct to a Terraform types.Object.
func PermissionsFromAPI(perms APIKeyPermissions) types.Object {
	return types.ObjectValueMust(PermissionsAttrTypes(), map[string]attr.Value{
		"monitor":                  ResourcePermFromAPI(perms.Monitor),
		"monitor_group":            ResourcePermFromAPI(perms.MonitorGroup),
		"tag":                      ResourcePermFromAPI(perms.Tag),
		"notification_channel":     ResourcePermFromAPI(perms.NotificationChannel),
		"maintenance_window":       ResourcePermFromAPI(perms.MaintenanceWindow),
		"status_page":              ResourcePermFromAPI(perms.StatusPage),
		"status_page_announcement": ResourcePermFromAPI(perms.StatusPageAnnouncement),
		"incident":                 ResourcePermFromAPI(perms.Incident),
		"incident_message":         ResourcePermFromAPI(perms.IncidentMessage),
		"agent_node":               ResourcePermFromAPI(perms.AgentNode),
		"scheduled_report":         ResourcePermFromAPI(perms.ScheduledReport),
		"manual_report":            ResourcePermFromAPI(perms.ManualReport),
		"sla_policy":               ResourcePermFromAPI(perms.SlaPolicy),
		"api_key":                  ResourcePermFromAPI(perms.APIKey),
	})
}
