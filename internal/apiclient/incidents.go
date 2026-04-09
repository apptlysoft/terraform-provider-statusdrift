package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type Incident struct {
	ID         int     `json:"id"`
	MonitorID  int     `json:"monitorId"`
	Message    *string `json:"message"`
	Priority   *string `json:"priority"`
	Resolved   bool    `json:"resolved"`
	ResolvedAt *string `json:"resolvedAt"`
	CreatedAt  string  `json:"createdAt"`
}

type IncidentInput struct {
	MonitorID int     `json:"monitorId"`
	Message   *string `json:"message,omitempty"`
	Priority  *string `json:"priority,omitempty"`
}

func (c *Client) CreateIncident(input IncidentInput) (*Incident, error) {
	body, err := c.post("/api/v1/incident", input)
	if err != nil {
		return nil, err
	}
	var result Incident
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling incident: %w", err)
	}
	return &result, nil
}

func (c *Client) GetIncident(id int) (*Incident, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/incident/%d", id))
	if err != nil {
		return nil, err
	}
	var result Incident
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling incident: %w", err)
	}
	return &result, nil
}

func (c *Client) ResolveIncident(id int) (*Incident, error) {
	body, err := c.post(fmt.Sprintf("/api/v1/incident/%d/resolve", id), nil)
	if err != nil {
		return nil, err
	}
	var result Incident
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling incident: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteIncident(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/incident/%d", id))
}

func (c *Client) ListIncidents() ([]Incident, error) {
	return AutoPaginate[Incident](c, "/api/v1/incident")
}

// IncidentListFilters contains optional filters for listing incidents.
type IncidentListFilters struct {
	Status    *string // "open" or "resolved"
	Type      *string // "manual" or "system"
	Priority  *string // "low", "medium", "high", "critical"
	MonitorID *int
}

func (c *Client) ListIncidentsFiltered(filters IncidentListFilters) ([]Incident, error) {
	params := url.Values{}
	if filters.Status != nil {
		params.Set("status", *filters.Status)
	}
	if filters.Type != nil {
		params.Set("type", *filters.Type)
	}
	if filters.Priority != nil {
		params.Set("priority", *filters.Priority)
	}
	if filters.MonitorID != nil {
		params.Set("monitorId", strconv.Itoa(*filters.MonitorID))
	}

	path := "/api/v1/incident"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return AutoPaginate[Incident](c, path)
}
