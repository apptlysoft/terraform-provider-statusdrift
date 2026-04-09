package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type NotificationChannel struct {
	ID      int                    `json:"id"`
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

type NotificationChannelInput struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

func (c *Client) CreateNotificationChannel(input NotificationChannelInput) (*NotificationChannel, error) {
	body, err := c.post("/api/v1/notification-channel", input)
	if err != nil {
		return nil, err
	}
	var result NotificationChannel
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling notification channel: %w", err)
	}
	return &result, nil
}

func (c *Client) GetNotificationChannel(id int) (*NotificationChannel, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/notification-channel/%d", id))
	if err != nil {
		return nil, err
	}
	var result NotificationChannel
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling notification channel: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateNotificationChannel(id int, input NotificationChannelInput) (*NotificationChannel, error) {
	body, err := c.put(fmt.Sprintf("/api/v1/notification-channel/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result NotificationChannel
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling notification channel: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteNotificationChannel(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/notification-channel/%d", id))
}

func (c *Client) ListNotificationChannels() ([]NotificationChannel, error) {
	return AutoPaginate[NotificationChannel](c, "/api/v1/notification-channel")
}

func (c *Client) ListNotificationChannelsWithSearch(search string) ([]NotificationChannel, error) {
	return AutoPaginate[NotificationChannel](c, "/api/v1/notification-channel?search="+url.QueryEscape(search))
}
