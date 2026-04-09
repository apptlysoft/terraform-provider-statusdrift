package apiclient

import (
	"encoding/json"
	"fmt"
)

type StatusPageSection struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Position    *int    `json:"position"`
	Expanded    bool    `json:"expanded"`
	Enabled     bool    `json:"enabled"`
}

type StatusPageSectionInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Position    *int    `json:"position,omitempty"`
	Expanded    bool    `json:"expanded"`
	Enabled     bool    `json:"enabled"`
}

func (c *Client) CreateStatusPageSection(statusPageID int, input StatusPageSectionInput) (*StatusPageSection, error) {
	body, err := c.post(fmt.Sprintf("/api/v1/status-page/%d/section", statusPageID), input)
	if err != nil {
		return nil, err
	}
	var result StatusPageSection
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page section: %w", err)
	}
	return &result, nil
}

func (c *Client) GetStatusPageSection(statusPageID, id int) (*StatusPageSection, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/status-page/%d/section/%d", statusPageID, id))
	if err != nil {
		return nil, err
	}
	var result StatusPageSection
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page section: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateStatusPageSection(statusPageID, id int, input StatusPageSectionInput) (*StatusPageSection, error) {
	body, err := c.put(fmt.Sprintf("/api/v1/status-page/%d/section/%d", statusPageID, id), input)
	if err != nil {
		return nil, err
	}
	var result StatusPageSection
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page section: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteStatusPageSection(statusPageID, id int) error {
	return c.delete(fmt.Sprintf("/api/v1/status-page/%d/section/%d", statusPageID, id))
}

func (c *Client) ListStatusPageSections(statusPageID int) ([]StatusPageSection, error) {
	return AutoPaginate[StatusPageSection](c, fmt.Sprintf("/api/v1/status-page/%d/section", statusPageID))
}
