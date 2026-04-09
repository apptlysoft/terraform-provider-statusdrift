package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type StatusPage struct {
	ID                   int          `json:"id"`
	Slug                 string       `json:"slug"`
	Name                 string       `json:"name"`
	Type                 string       `json:"type"`
	Enabled              bool         `json:"enabled"`
	Description          *string      `json:"description"`
	MonitorID            *int         `json:"monitorId"`
	MonitorGroupID       *int         `json:"monitorGroupId"`
	TagID                *int         `json:"tagId"`
	ShowAnnouncements    bool         `json:"showAnnouncements"`
	ShowIncidents        bool         `json:"showIncidents"`
	IncidentDisplayTypes []string     `json:"incidentDisplayTypes"`
	ShowHistory          bool         `json:"showHistory"`
	HistoryDays          int          `json:"historyDays"`
	ShowUptime           bool         `json:"showUptime"`
	ShowResponseTime     bool         `json:"showResponseTime"`
	ShowOverallStatus    bool         `json:"showOverallStatus"`
	LogoURL              *string      `json:"logoUrl"`
	FaviconURL           *string      `json:"faviconUrl"`
	CustomDomain         *string      `json:"customDomain"`
	PrimaryColor         *string      `json:"primaryColor"`
	CustomCSS            *string      `json:"customCss"`
	HidePoweredBy        bool         `json:"hidePoweredBy"`
	Password             *string      `json:"password"`
	Timezone             *string      `json:"timezone"`
	GoogleAnalyticsID    *string      `json:"googleAnalyticsId"`
	SocialLinks          []SocialLink `json:"socialLinks"`
	ColorTheme           string       `json:"colorTheme"`
}

type StatusPageInput struct {
	Name                 string       `json:"name"`
	Type                 string       `json:"type"`
	Enabled              bool         `json:"enabled"`
	Description          *string      `json:"description,omitempty"`
	MonitorID            *int         `json:"monitorId,omitempty"`
	MonitorGroupID       *int         `json:"monitorGroupId,omitempty"`
	TagID                *int         `json:"tagId,omitempty"`
	ShowAnnouncements    bool         `json:"showAnnouncements"`
	ShowIncidents        bool         `json:"showIncidents"`
	IncidentDisplayTypes []string     `json:"incidentDisplayTypes,omitempty"`
	ShowHistory          bool         `json:"showHistory"`
	HistoryDays          int          `json:"historyDays"`
	ShowUptime           bool         `json:"showUptime"`
	ShowResponseTime     bool         `json:"showResponseTime"`
	ShowOverallStatus    bool         `json:"showOverallStatus"`
	LogoURL              *string      `json:"logoUrl,omitempty"`
	FaviconURL           *string      `json:"faviconUrl,omitempty"`
	CustomDomain         *string      `json:"customDomain,omitempty"`
	PrimaryColor         *string      `json:"primaryColor,omitempty"`
	CustomCSS            *string      `json:"customCss,omitempty"`
	HidePoweredBy        bool         `json:"hidePoweredBy"`
	Password             *string      `json:"password,omitempty"`
	Timezone             *string      `json:"timezone,omitempty"`
	GoogleAnalyticsID    *string      `json:"googleAnalyticsId,omitempty"`
	SocialLinks          []SocialLink `json:"socialLinks,omitempty"`
	ColorTheme           string       `json:"colorTheme"`
}

func (c *Client) CreateStatusPage(input StatusPageInput) (*StatusPage, error) {
	body, err := c.post("/api/v1/status-page", input)
	if err != nil {
		return nil, err
	}
	var result StatusPage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page: %w", err)
	}
	return &result, nil
}

func (c *Client) GetStatusPage(id int) (*StatusPage, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/status-page/%d", id))
	if err != nil {
		return nil, err
	}
	var result StatusPage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateStatusPage(id int, input StatusPageInput) (*StatusPage, error) {
	body, err := c.put(fmt.Sprintf("/api/v1/status-page/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result StatusPage
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling status page: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteStatusPage(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/status-page/%d", id))
}

func (c *Client) ListStatusPages() ([]StatusPage, error) {
	return AutoPaginate[StatusPage](c, "/api/v1/status-page")
}

func (c *Client) ListStatusPagesWithSearch(search string) ([]StatusPage, error) {
	return AutoPaginate[StatusPage](c, "/api/v1/status-page?search="+url.QueryEscape(search))
}
