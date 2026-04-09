package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Tag struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

type TagInput struct {
	Name  string  `json:"name"`
	Color *string `json:"color,omitempty"`
}

func (c *Client) CreateTag(input TagInput) (*Tag, error) {
	body, err := c.post("/api/v1/tag", input)
	if err != nil {
		return nil, err
	}
	var result Tag
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling tag: %w", err)
	}
	return &result, nil
}

func (c *Client) GetTag(id int) (*Tag, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/tag/%d", id))
	if err != nil {
		return nil, err
	}
	var result Tag
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling tag: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateTag(id int, input TagInput) (*Tag, error) {
	body, err := c.put(fmt.Sprintf("/api/v1/tag/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result Tag
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling tag: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteTag(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/tag/%d", id))
}

func (c *Client) ListTags() ([]Tag, error) {
	return AutoPaginate[Tag](c, "/api/v1/tag")
}

func (c *Client) ListTagsWithSearch(search string) ([]Tag, error) {
	return AutoPaginate[Tag](c, "/api/v1/tag?search="+url.QueryEscape(search))
}
