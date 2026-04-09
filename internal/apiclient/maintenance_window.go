package apiclient

import (
	"encoding/json"
	"fmt"
)

type MaintenanceWindow struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Repeat         string   `json:"repeat"`
	Enabled        bool     `json:"enabled"`
	TargetType     string   `json:"targetType"`
	TargetID       *int     `json:"targetId"`
	StartDateTime  *string  `json:"startDateTime"`
	EndDateTime    *string  `json:"endDateTime"`
	StartTime      *string  `json:"startTime"`
	EndTime        *string  `json:"endTime"`
	Timezone       string   `json:"timezone"`
	WeeklyDays     []string `json:"weeklyDays"`
	MonthlyDayType *string  `json:"monthlyDayType"`
	MonthlyDay     *int     `json:"monthlyDay"`
}

type MaintenanceWindowInput struct {
	Name           string   `json:"name"`
	Repeat         string   `json:"repeat"`
	Enabled        *bool    `json:"enabled,omitempty"`
	TargetType     string   `json:"targetType"`
	TargetID       *int     `json:"targetId,omitempty"`
	StartDateTime  *string  `json:"startDateTime,omitempty"`
	EndDateTime    *string  `json:"endDateTime,omitempty"`
	StartTime      *string  `json:"startTime,omitempty"`
	EndTime        *string  `json:"endTime,omitempty"`
	Timezone       string   `json:"timezone,omitempty"`
	WeeklyDays     []string `json:"weeklyDays,omitempty"`
	MonthlyDayType *string  `json:"monthlyDayType,omitempty"`
	MonthlyDay     *int     `json:"monthlyDay,omitempty"`
}

// MaintenanceWindowPatch is used for PATCH updates — only changed fields are sent.
type MaintenanceWindowPatch map[string]interface{}

func (c *Client) CreateMaintenanceWindow(input MaintenanceWindowInput) (*MaintenanceWindow, error) {
	body, err := c.post("/api/v1/maintenance-windows", input)
	if err != nil {
		return nil, err
	}
	var result MaintenanceWindow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling maintenance window: %w", err)
	}
	return &result, nil
}

func (c *Client) GetMaintenanceWindow(id int) (*MaintenanceWindow, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/maintenance-windows/%d", id))
	if err != nil {
		return nil, err
	}
	var result MaintenanceWindow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling maintenance window: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateMaintenanceWindow(id int, patch MaintenanceWindowPatch) (*MaintenanceWindow, error) {
	body, err := c.patch(fmt.Sprintf("/api/v1/maintenance-windows/%d", id), patch)
	if err != nil {
		return nil, err
	}
	var result MaintenanceWindow
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling maintenance window: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteMaintenanceWindow(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/maintenance-windows/%d", id))
}

func (c *Client) ListMaintenanceWindows() ([]MaintenanceWindow, error) {
	return AutoPaginate[MaintenanceWindow](c, "/api/v1/maintenance-windows")
}
