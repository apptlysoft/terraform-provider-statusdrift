package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Collaborator represents an organization member/collaborator returned by the API.
type Collaborator struct {
	ID     int     `json:"id"`
	Email  string  `json:"collaboratorEmail"`
	Name   *string `json:"collaboratorName"`
	Role   string  `json:"role"`
	Status string  `json:"status"`
}

// collaboratorListResponse matches the API's response shape (items/pages, not data/totalPages).
type collaboratorListResponse struct {
	Items []Collaborator `json:"items"`
	Total int            `json:"total"`
	Pages int            `json:"pages"`
}

func (c *Client) GetCollaborator(id int) (*Collaborator, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/collaborator/%d", id))
	if err != nil {
		return nil, err
	}
	var result Collaborator
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling collaborator: %w", err)
	}
	return &result, nil
}

func (c *Client) listCollaboratorsPaginated(basePath string) ([]Collaborator, error) {
	var all []Collaborator
	page := 0

	for {
		separator := "?"
		if strings.Contains(basePath, "?") {
			separator = "&"
		}
		path := fmt.Sprintf("%s%spage=%d&perPage=100", basePath, separator, page)

		body, err := c.get(path)
		if err != nil {
			return nil, err
		}

		var resp collaboratorListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("unmarshaling collaborator list: %w", err)
		}

		all = append(all, resp.Items...)

		if page+1 >= resp.Pages || len(resp.Items) == 0 {
			break
		}
		page++
	}

	return all, nil
}

func (c *Client) ListCollaborators() ([]Collaborator, error) {
	return c.listCollaboratorsPaginated("/api/v1/collaborator?includeOwner=true")
}

func (c *Client) ListCollaboratorsWithSearch(search string) ([]Collaborator, error) {
	return c.listCollaboratorsPaginated("/api/v1/collaborator?includeOwner=true&search=" + url.QueryEscape(search))
}
