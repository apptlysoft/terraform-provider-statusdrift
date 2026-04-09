package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type SlaPolicyMonitorRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (r SlaPolicyMonitorRef) GetID() int { return r.ID }

type SlaPolicyTagRef struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

func (r SlaPolicyTagRef) GetID() int { return r.ID }

type SlaPolicyGroupRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (r SlaPolicyGroupRef) GetID() int { return r.ID }

type SlaPolicy struct {
	ID                      int                   `json:"id"`
	Name                    string                `json:"name"`
	UptimeTarget            float64               `json:"uptimeTarget"`
	WindowType              string                `json:"windowType"`
	ResponseTimeSlaEnabled  bool                  `json:"responseTimeSlaEnabled"`
	ResponseTimeThresholdMs *int                  `json:"responseTimeThresholdMs"`
	Enabled                 bool                  `json:"enabled"`
	Monitors                []SlaPolicyMonitorRef `json:"monitors"`
	Tags                    []SlaPolicyTagRef     `json:"tags"`
	Groups                  []SlaPolicyGroupRef   `json:"groups"`
	CreatedAt               string                `json:"createdAt"`
	UpdatedAt               string                `json:"updatedAt"`
}

type SlaPolicyCreateInput struct {
	Name                    string  `json:"name"`
	UptimeTarget            float64 `json:"uptimeTarget"`
	WindowType              string  `json:"windowType"`
	ResponseTimeSlaEnabled  *bool   `json:"responseTimeSlaEnabled,omitempty"`
	ResponseTimeThresholdMs *int    `json:"responseTimeThresholdMs,omitempty"`
	Enabled                 *bool   `json:"enabled,omitempty"`
	MonitorIDs              []int   `json:"monitorIds,omitempty"`
	TagIDs                  []int   `json:"tagIds,omitempty"`
	GroupIDs                []int   `json:"groupIds,omitempty"`
}

type SlaPolicyUpdateInput struct {
	Name                    *string  `json:"name,omitempty"`
	UptimeTarget            *float64 `json:"uptimeTarget,omitempty"`
	WindowType              *string  `json:"windowType,omitempty"`
	ResponseTimeSlaEnabled  *bool    `json:"responseTimeSlaEnabled,omitempty"`
	ResponseTimeThresholdMs *int     `json:"responseTimeThresholdMs,omitempty"`
	Enabled                 *bool    `json:"enabled,omitempty"`
	MonitorIDs              []int    `json:"monitorIds,omitempty"`
	TagIDs                  []int    `json:"tagIds,omitempty"`
	GroupIDs                []int    `json:"groupIds,omitempty"`
}

func (c *Client) CreateSlaPolicy(input SlaPolicyCreateInput) (*SlaPolicy, error) {
	body, err := c.post("/api/v1/sla-policies", input)
	if err != nil {
		return nil, err
	}
	var result SlaPolicy
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling sla policy: %w", err)
	}
	return &result, nil
}

func (c *Client) GetSlaPolicy(id int) (*SlaPolicy, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/sla-policies/%d", id))
	if err != nil {
		return nil, err
	}
	var result SlaPolicy
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling sla policy: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateSlaPolicy(id int, input SlaPolicyUpdateInput) (*SlaPolicy, error) {
	body, err := c.patch(fmt.Sprintf("/api/v1/sla-policies/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result SlaPolicy
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling sla policy: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteSlaPolicy(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/sla-policies/%d", id))
}

func (c *Client) ListSlaPolicies() ([]SlaPolicy, error) {
	return AutoPaginate[SlaPolicy](c, "/api/v1/sla-policies")
}

func (c *Client) ListSlaPoliciesWithSearch(search string) ([]SlaPolicy, error) {
	return AutoPaginate[SlaPolicy](c, "/api/v1/sla-policies?search="+url.QueryEscape(search))
}
