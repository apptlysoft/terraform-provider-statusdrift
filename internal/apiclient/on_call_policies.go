package apiclient

import (
	"encoding/json"
	"fmt"
)

// OnCallPolicyStepScheduleRef is the schedule reference in a policy step response.
type OnCallPolicyStepScheduleRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// OnCallPolicyStepRef represents an escalation step in a policy response.
type OnCallPolicyStepRef struct {
	ID                 *int                         `json:"id"`
	Position           int                          `json:"position"`
	TargetType         string                       `json:"targetType"`
	DelayMinutes       int                          `json:"delayMinutes"`
	Schedule           *OnCallPolicyStepScheduleRef `json:"schedule"`
	OrganizationMember *OnCallMember                `json:"organizationMember"`
}

// OnCallPolicyMonitorGroupRef is a monitor group reference in a policy response.
type OnCallPolicyMonitorGroupRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// OnCallPolicy represents an on-call escalation policy returned by the API.
type OnCallPolicy struct {
	ID           int                          `json:"id"`
	Name         string                       `json:"name"`
	Enabled      bool                         `json:"enabled"`
	IsDefault    bool                         `json:"isDefault"`
	ScopeType    string                       `json:"scopeType"`
	MonitorGroup *OnCallPolicyMonitorGroupRef `json:"monitorGroup"`
	CanManage    bool                         `json:"canManage"`
	Steps        []OnCallPolicyStepRef        `json:"steps"`
}

// OnCallPolicyStepInput is a single step in a policy create/update request.
type OnCallPolicyStepInput struct {
	TargetType           string `json:"targetType"`
	ScheduleID           *int   `json:"scheduleId,omitempty"`
	OrganizationMemberID *int   `json:"organizationMemberId,omitempty"`
	DelayMinutes         int    `json:"delayMinutes"`
}

// OnCallPolicyInput is the request body for creating/updating an on-call policy.
type OnCallPolicyInput struct {
	Name           string                  `json:"name"`
	Enabled        bool                    `json:"enabled"`
	ScopeType      *string                 `json:"scopeType,omitempty"`
	MonitorGroupID *int                    `json:"monitorGroupId,omitempty"`
	Steps          []OnCallPolicyStepInput `json:"steps"`
}

// ListOnCallPolicies fetches all on-call policies. The endpoint returns a plain array.
func (c *Client) ListOnCallPolicies() ([]OnCallPolicy, error) {
	body, err := c.get("/api/v1/on-call/policies")
	if err != nil {
		return nil, err
	}
	var result []OnCallPolicy
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call policies: %w", err)
	}
	return result, nil
}

// GetOnCallPolicy fetches a single on-call policy by ID.
// Since there is no individual GET endpoint, this lists all policies and filters.
func (c *Client) GetOnCallPolicy(id int) (*OnCallPolicy, error) {
	policies, err := c.ListOnCallPolicies()
	if err != nil {
		return nil, err
	}
	for _, p := range policies {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("on-call policy %d not found", id)}
}

func (c *Client) CreateOnCallPolicy(input OnCallPolicyInput) (*OnCallPolicy, error) {
	body, err := c.post("/api/v1/on-call/policies", input)
	if err != nil {
		return nil, err
	}
	var result OnCallPolicy
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call policy: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateOnCallPolicy(id int, input OnCallPolicyInput) (*OnCallPolicy, error) {
	body, err := c.patch(fmt.Sprintf("/api/v1/on-call/policies/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result OnCallPolicy
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call policy: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteOnCallPolicy(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/on-call/policies/%d", id))
}

// SetDefaultOnCallPolicy marks a policy as the organization default.
func (c *Client) SetDefaultOnCallPolicy(id int) (*OnCallPolicy, error) {
	body, err := c.post(fmt.Sprintf("/api/v1/on-call/policies/%d/set-default", id), nil)
	if err != nil {
		return nil, err
	}
	var result OnCallPolicy
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call policy: %w", err)
	}
	return &result, nil
}
