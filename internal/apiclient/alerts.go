package apiclient

import (
	"encoding/json"
	"fmt"
)

// Alert represents a standalone alert from the read-only API.
type Alert struct {
	ID                       int      `json:"id"`
	MonitorID                int      `json:"monitorId"`
	NotificationChannelIDs   []int    `json:"notificationChannelIds"`
	Conditions               []string `json:"conditions"`
	DelayMinutes             *int     `json:"delayMinutes"`
	Enabled                  bool     `json:"enabled"`
	RecurringIntervalMinutes *int     `json:"recurringIntervalMinutes"`
	MaxRecurrences           *int     `json:"maxRecurrences"`
	NotifyOnCall             bool     `json:"notifyOnCall"`
	OnCallPolicyID           *int     `json:"onCallPolicyId"`
}

func (c *Client) GetAlert(id int) (*Alert, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/alert/%d", id))
	if err != nil {
		return nil, err
	}
	var result Alert
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling alert: %w", err)
	}
	return &result, nil
}

func (c *Client) ListAlerts() ([]Alert, error) {
	return AutoPaginate[Alert](c, "/api/v1/alert")
}
