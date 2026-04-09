package apiclient

import (
	"encoding/json"
	"fmt"
)

// OnCallMember represents an organization member in on-call context.
type OnCallMember struct {
	ID     int     `json:"id"`
	Email  string  `json:"email"`
	Name   *string `json:"name"`
	Role   string  `json:"role"`
	Status string  `json:"status"`
}

// OnCallParticipant represents a participant in an on-call schedule rotation.
type OnCallParticipant struct {
	ID                 int          `json:"id"`
	Position           int          `json:"position"`
	OrganizationMember OnCallMember `json:"organizationMember"`
}

// OnCallOverrideResponse represents an override returned within a schedule response.
type OnCallOverrideResponse struct {
	ID                int          `json:"id"`
	StartsAt          string       `json:"startsAt"`
	EndsAt            string       `json:"endsAt"`
	Note              *string      `json:"note"`
	ReplacementMember OnCallMember `json:"replacementMember"`
}

// OnCallSchedule represents an on-call schedule returned by the API.
type OnCallSchedule struct {
	ID                     int                      `json:"id"`
	Name                   string                   `json:"name"`
	Timezone               string                   `json:"timezone"`
	Frequency              string                   `json:"frequency"`
	RotationInterval       int                      `json:"rotationInterval"`
	StartsAt               string                   `json:"startsAt"`
	Enabled                bool                     `json:"enabled"`
	Participants           []OnCallParticipant      `json:"participants"`
	Overrides              []OnCallOverrideResponse `json:"overrides"`
	CurrentResponder       *OnCallMember            `json:"currentResponder"`
	CurrentResponderSource *string                  `json:"currentResponderSource"`
	NextHandoffAt          *string                  `json:"nextHandoffAt"`
}

// OnCallScheduleInput is the request body for creating/updating an on-call schedule.
type OnCallScheduleInput struct {
	Name             string `json:"name"`
	Timezone         string `json:"timezone"`
	Frequency        string `json:"frequency"`
	RotationInterval int    `json:"rotationInterval"`
	StartsAt         string `json:"startsAt"`
	Enabled          bool   `json:"enabled"`
	ParticipantIDs   []int  `json:"participantIds"`
}

// OnCallOverrideInput is the request body for creating/updating an on-call override.
type OnCallOverrideInput struct {
	ReplacementMemberID int     `json:"replacementMemberId"`
	StartsAt            string  `json:"startsAt"`
	EndsAt              string  `json:"endsAt"`
	Note                *string `json:"note,omitempty"`
}

// ListOnCallSchedules fetches all on-call schedules. The endpoint returns a plain array.
func (c *Client) ListOnCallSchedules() ([]OnCallSchedule, error) {
	body, err := c.get("/api/v1/on-call/schedules")
	if err != nil {
		return nil, err
	}
	var result []OnCallSchedule
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call schedules: %w", err)
	}
	return result, nil
}

// GetOnCallSchedule fetches a single on-call schedule by ID.
// Since there is no individual GET endpoint, this lists all schedules and filters.
func (c *Client) GetOnCallSchedule(id int) (*OnCallSchedule, error) {
	schedules, err := c.ListOnCallSchedules()
	if err != nil {
		return nil, err
	}
	for _, s := range schedules {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("on-call schedule %d not found", id)}
}

func (c *Client) CreateOnCallSchedule(input OnCallScheduleInput) (*OnCallSchedule, error) {
	body, err := c.post("/api/v1/on-call/schedules", input)
	if err != nil {
		return nil, err
	}
	var result OnCallSchedule
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call schedule: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateOnCallSchedule(id int, input OnCallScheduleInput) (*OnCallSchedule, error) {
	body, err := c.patch(fmt.Sprintf("/api/v1/on-call/schedules/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result OnCallSchedule
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call schedule: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteOnCallSchedule(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/on-call/schedules/%d", id))
}

// CreateOnCallOverride creates an override on a schedule.
func (c *Client) CreateOnCallOverride(scheduleID int, input OnCallOverrideInput) (*OnCallOverrideResponse, error) {
	body, err := c.post(fmt.Sprintf("/api/v1/on-call/schedules/%d/overrides", scheduleID), input)
	if err != nil {
		return nil, err
	}
	var result OnCallOverrideResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call override: %w", err)
	}
	return &result, nil
}

// UpdateOnCallOverride updates an override on a schedule.
func (c *Client) UpdateOnCallOverride(scheduleID, overrideID int, input OnCallOverrideInput) (*OnCallOverrideResponse, error) {
	body, err := c.patch(fmt.Sprintf("/api/v1/on-call/schedules/%d/overrides/%d", scheduleID, overrideID), input)
	if err != nil {
		return nil, err
	}
	var result OnCallOverrideResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling on-call override: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteOnCallOverride(scheduleID, overrideID int) error {
	return c.delete(fmt.Sprintf("/api/v1/on-call/schedules/%d/overrides/%d", scheduleID, overrideID))
}

// GetOnCallOverride fetches a single override by listing all schedules and searching.
func (c *Client) GetOnCallOverride(scheduleID, overrideID int) (*OnCallOverrideResponse, error) {
	schedule, err := c.GetOnCallSchedule(scheduleID)
	if err != nil {
		return nil, err
	}
	for _, o := range schedule.Overrides {
		if o.ID == overrideID {
			return &o, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("on-call override %d not found on schedule %d", overrideID, scheduleID)}
}
