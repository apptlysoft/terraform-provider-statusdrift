package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the StatusDrift API client.
type Client struct {
	BaseURL    string
	APIKey     string
	UserAgent  string
	HTTPClient *http.Client
}

// NewClient creates a new StatusDrift API client.
func NewClient(baseURL, apiKey, version string) *Client {
	return &Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		APIKey:    apiKey,
		UserAgent: fmt.Sprintf("terraform-provider-statusdrift/%s", version),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PaginatedResponse represents a paginated API response.
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	TotalPages int `json:"totalPages"`
}

func (c *Client) doRequest(method, path string, body any) ([]byte, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

func parseAPIError(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode}

	var errResp struct {
		Message    string           `json:"message"`
		Detail     string           `json:"detail"`
		Violations []FieldViolation `json:"violations"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil {
		apiErr.Message = errResp.Message
		if apiErr.Message == "" {
			apiErr.Message = errResp.Detail
		}
		apiErr.Violations = errResp.Violations
	}

	if apiErr.Message == "" {
		switch statusCode {
		case 401:
			apiErr.Message = "unauthorized — check your API key"
		case 403:
			apiErr.Message = "forbidden — your API key does not have permission for this action"
		case 404:
			apiErr.Message = "resource not found"
		case 409:
			apiErr.Message = "conflict"
		default:
			apiErr.Message = string(body)
		}
	}

	return apiErr
}

// get performs a GET request and returns the raw response body.
func (c *Client) get(path string) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// post performs a POST request and returns the raw response body.
func (c *Client) post(path string, body any) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, body)
}

// put performs a PUT request and returns the raw response body.
func (c *Client) put(path string, body any) ([]byte, error) {
	return c.doRequest(http.MethodPut, path, body)
}

// patch performs a PATCH request and returns the raw response body.
func (c *Client) patch(path string, body any) ([]byte, error) {
	return c.doRequest(http.MethodPatch, path, body)
}

// delete performs a DELETE request.
func (c *Client) delete(path string) error {
	_, err := c.doRequest(http.MethodDelete, path, nil)
	return err
}

// AutoPaginate fetches all pages of a paginated endpoint.
func AutoPaginate[T any](c *Client, basePath string) ([]T, error) {
	var all []T
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

		var resp PaginatedResponse[T]
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("unmarshaling paginated response: %w", err)
		}

		all = append(all, resp.Data...)

		if page+1 >= resp.TotalPages || len(resp.Data) == 0 {
			break
		}
		page++
	}

	return all, nil
}
