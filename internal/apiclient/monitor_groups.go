package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type MonitorGroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MonitorGroupInput struct {
	Name string `json:"name"`
}

func (c *Client) CreateMonitorGroup(input MonitorGroupInput) (*MonitorGroup, error) {
	body, err := c.post("/api/v1/monitor-group", input)
	if err != nil {
		return nil, err
	}
	var result MonitorGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling monitor group: %w", err)
	}
	return &result, nil
}

func (c *Client) GetMonitorGroup(id int) (*MonitorGroup, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/monitor-group/%d", id))
	if err != nil {
		return nil, err
	}
	var result MonitorGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling monitor group: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateMonitorGroup(id int, input MonitorGroupInput) (*MonitorGroup, error) {
	body, err := c.put(fmt.Sprintf("/api/v1/monitor-group/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result MonitorGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling monitor group: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteMonitorGroup(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/monitor-group/%d", id))
}

func (c *Client) ListMonitorGroups() ([]MonitorGroup, error) {
	return AutoPaginate[MonitorGroup](c, "/api/v1/monitor-group")
}

func (c *Client) ListMonitorGroupsWithSearch(search string) ([]MonitorGroup, error) {
	return AutoPaginate[MonitorGroup](c, "/api/v1/monitor-group?search="+url.QueryEscape(search))
}
