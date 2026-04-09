package apiclient

import (
	"encoding/json"
	"fmt"
)

type StatusPageAnnouncement struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	Type     string  `json:"type"`
	Enabled  bool    `json:"enabled"`
	StartsAt *string `json:"startsAt"`
	EndsAt   *string `json:"endsAt"`
}

type StatusPageAnnouncementInput struct {
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	Type     string  `json:"type"`
	Enabled  bool    `json:"enabled"`
	StartsAt *string `json:"startsAt,omitempty"`
	EndsAt   *string `json:"endsAt,omitempty"`
}

func (c *Client) CreateStatusPageAnnouncement(statusPageID int, input StatusPageAnnouncementInput) (*StatusPageAnnouncement, error) {
	body, err := c.post(fmt.Sprintf("/api/v1/status-page/%d/announcement", statusPageID), input)
	if err != nil {
		return nil, err
	}
	var result StatusPageAnnouncement
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page announcement: %w", err)
	}
	return &result, nil
}

func (c *Client) GetStatusPageAnnouncement(statusPageID, id int) (*StatusPageAnnouncement, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/status-page/%d/announcement/%d", statusPageID, id))
	if err != nil {
		return nil, err
	}
	var result StatusPageAnnouncement
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page announcement: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateStatusPageAnnouncement(statusPageID, id int, input StatusPageAnnouncementInput) (*StatusPageAnnouncement, error) {
	body, err := c.put(fmt.Sprintf("/api/v1/status-page/%d/announcement/%d", statusPageID, id), input)
	if err != nil {
		return nil, err
	}
	var result StatusPageAnnouncement
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page announcement: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteStatusPageAnnouncement(statusPageID, id int) error {
	return c.delete(fmt.Sprintf("/api/v1/status-page/%d/announcement/%d", statusPageID, id))
}

func (c *Client) ListStatusPageAnnouncements(statusPageID int) ([]StatusPageAnnouncement, error) {
	return AutoPaginate[StatusPageAnnouncement](c, fmt.Sprintf("/api/v1/status-page/%d/announcement", statusPageID))
}
