package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ResourcePermissions represents CRUD permissions for a single resource type.
type ResourcePermissions struct {
	Create bool `json:"create"`
	Read   bool `json:"read"`
	Update bool `json:"update"`
	Delete bool `json:"delete"`
}

// APIKeyPermissions represents the full permissions map for an API key.
type APIKeyPermissions struct {
	Monitor                *ResourcePermissions `json:"monitor,omitempty"`
	MonitorGroup           *ResourcePermissions `json:"monitorGroup,omitempty"`
	Tag                    *ResourcePermissions `json:"tag,omitempty"`
	NotificationChannel    *ResourcePermissions `json:"notificationChannel,omitempty"`
	MaintenanceWindow      *ResourcePermissions `json:"maintenanceWindow,omitempty"`
	StatusPage             *ResourcePermissions `json:"statusPage,omitempty"`
	StatusPageAnnouncement *ResourcePermissions `json:"statusPageAnnouncement,omitempty"`
	Incident               *ResourcePermissions `json:"incident,omitempty"`
	IncidentMessage        *ResourcePermissions `json:"incidentMessage,omitempty"`
	AgentNode              *ResourcePermissions `json:"agentNode,omitempty"`
	ScheduledReport        *ResourcePermissions `json:"scheduledReport,omitempty"`
	ManualReport           *ResourcePermissions `json:"manualReport,omitempty"`
	SlaPolicy              *ResourcePermissions `json:"slaPolicy,omitempty"`
	APIKey                 *ResourcePermissions `json:"apiKey,omitempty"`
}

// APIKey represents an API key returned by the API.
type APIKey struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Key         string            `json:"key,omitempty"`
	KeyPrefix   string            `json:"keyPrefix"`
	Permissions APIKeyPermissions `json:"permissions"`
	ExpiresAt   *string           `json:"expiresAt"`
	LastUsedAt  *string           `json:"lastUsedAt"`
	CreatedAt   string            `json:"createdAt"`
	IsExpired   bool              `json:"isExpired"`
}

// APIKeyCreateInput is the request body for creating an API key.
type APIKeyCreateInput struct {
	Name          string            `json:"name"`
	Permissions   APIKeyPermissions `json:"permissions"`
	ExpiresInDays *int              `json:"expiresInDays,omitempty"`
}

// APIKeyUpdateInput is the request body for updating an API key.
type APIKeyUpdateInput struct {
	Name        *string            `json:"name,omitempty"`
	Permissions *APIKeyPermissions `json:"permissions,omitempty"`
}

func (c *Client) CreateAPIKey(input APIKeyCreateInput) (*APIKey, error) {
	body, err := c.post("/api/v1/api-key", input)
	if err != nil {
		return nil, err
	}
	var result APIKey
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling api key: %w", err)
	}
	return &result, nil
}

func (c *Client) GetAPIKey(id int) (*APIKey, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/api-key/%d", id))
	if err != nil {
		return nil, err
	}
	var result APIKey
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling api key: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateAPIKey(id int, input APIKeyUpdateInput) (*APIKey, error) {
	body, err := c.put(fmt.Sprintf("/api/v1/api-key/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result APIKey
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling api key: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteAPIKey(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/api-key/%d", id))
}

func (c *Client) ListAPIKeys() ([]APIKey, error) {
	return AutoPaginate[APIKey](c, "/api/v1/api-key")
}

func (c *Client) ListAPIKeysWithSearch(search string) ([]APIKey, error) {
	return AutoPaginate[APIKey](c, "/api/v1/api-key?search="+url.QueryEscape(search))
}
