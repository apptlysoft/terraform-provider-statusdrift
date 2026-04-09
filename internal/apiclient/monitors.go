package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// FlexInt handles JSON values that may be either a number or a string-encoded number.
type FlexInt int

func (fi *FlexInt) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("cannot convert %q to int: %w", s, err)
		}
		*fi = FlexInt(n)
		return nil
	}
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*fi = FlexInt(n)
	return nil
}

type Monitor struct {
	ID                   FlexInt                `json:"id"`
	Name                 string                 `json:"name"`
	Type                 string                 `json:"type"`
	Interval             int                    `json:"interval"`
	Enabled              bool                   `json:"enabled"`
	LocationType         string                 `json:"locationType"`
	LocationIDs          []int                  `json:"locationIds"`
	Region               *string                `json:"region"`
	URL                  *string                `json:"url"`
	Host                 *string                `json:"host"`
	Port                 *int                   `json:"port"`
	HTTPMethod           *string                `json:"httpMethod"`
	HTTPPostBody         *string                `json:"httpPostBody"`
	HTTPContentType      *string                `json:"httpContentType"`
	MaxRedirects         *int                   `json:"maxRedirects"`
	AllowInvalidCert     bool                   `json:"allowInvalidCertificate"`
	CacheBuster          bool                   `json:"cacheBuster"`
	HTTPAuth             bool                   `json:"httpAuth"`
	HTTPUsername         *string                `json:"httpUsername"`
	HTTPPassword         *string                `json:"httpPassword"`
	RWTimeout            *int                   `json:"rwTimeout"`
	ConnectionTimeout    *int                   `json:"connectionTimeout"`
	ExpectedStatusCode   *string                `json:"expectedStatusCode"`
	Keyword              *string                `json:"keyword"`
	KeywordNegation      bool                   `json:"keywordNegation"`
	MonitorGroupID       *int                   `json:"monitorGroupId"`
	TagIDs               []int                  `json:"tagIds"`
	Alerts               []AlertConfig          `json:"alerts"`
	LocationsDownToAlert int                    `json:"locationsDownToAlert"`
	ChecksDownToAlert    int                    `json:"checksDownToAlert"`
	WarningThresholdMs   int                    `json:"warningThresholdMs"`
	DNSHostname          *string                `json:"dnsHostname"`
	DNSServers           []string               `json:"dnsServers"`
	DNSRecords           []DNSRecordConfig      `json:"dnsRecords"`
	DNSMonitoringMode    *string                `json:"dnsMonitoringMode"`
	Status               *string                `json:"status"`
	CreatedAt            *string                `json:"createdAt"`
	Extra                map[string]interface{} `json:"-"`
}

type AlertConfig struct {
	ID                       *int     `json:"id,omitempty"`
	NotificationChannelIDs   []int    `json:"notificationChannelIds"`
	Conditions               []string `json:"conditions"`
	DelayMinutes             *int     `json:"delayMinutes,omitempty"`
	Enabled                  bool     `json:"enabled"`
	RecurringIntervalMinutes *int     `json:"recurringIntervalMinutes,omitempty"`
	MaxRecurrences           *int     `json:"maxRecurrences,omitempty"`
	NotifyOnCall             bool     `json:"notifyOnCall"`
	OnCallPolicyID           *int     `json:"onCallPolicyId,omitempty"`
}

type DNSRecordConfig struct {
	RecordType     string   `json:"recordType"`
	Hostname       string   `json:"hostname"`
	ExpectedValues []string `json:"expectedValues"`
	MatchMode      string   `json:"matchMode"`
	Enabled        bool     `json:"enabled"`
}

type MonitorInput struct {
	Name                 string            `json:"name"`
	Type                 string            `json:"type"`
	Interval             int               `json:"interval"`
	Enabled              bool              `json:"enabled"`
	LocationType         string            `json:"locationType"`
	LocationIDs          []int             `json:"locationIds,omitempty"`
	Region               *string           `json:"region,omitempty"`
	URL                  *string           `json:"url,omitempty"`
	Host                 *string           `json:"host,omitempty"`
	Port                 *int              `json:"port,omitempty"`
	HTTPMethod           *string           `json:"httpMethod,omitempty"`
	HTTPPostBody         *string           `json:"httpPostBody,omitempty"`
	HTTPContentType      *string           `json:"httpContentType,omitempty"`
	MaxRedirects         *int              `json:"maxRedirects,omitempty"`
	AllowInvalidCert     bool              `json:"allowInvalidCertificate"`
	CacheBuster          bool              `json:"cacheBuster"`
	HTTPAuth             bool              `json:"httpAuth"`
	HTTPUsername         *string           `json:"httpUsername,omitempty"`
	HTTPPassword         *string           `json:"httpPassword,omitempty"`
	RWTimeout            *int              `json:"rwTimeout,omitempty"`
	ConnectionTimeout    *int              `json:"connectionTimeout,omitempty"`
	ExpectedStatusCode   *string           `json:"expectedStatusCode,omitempty"`
	Keyword              *string           `json:"keyword,omitempty"`
	KeywordNegation      bool              `json:"keywordNegation"`
	MonitorGroupID       *int              `json:"monitorGroupId,omitempty"`
	TagIDs               []int             `json:"tagIds,omitempty"`
	Alerts               []AlertConfig     `json:"alerts,omitempty"`
	LocationsDownToAlert int               `json:"locationsDownToAlert"`
	ChecksDownToAlert    int               `json:"checksDownToAlert"`
	WarningThresholdMs   int               `json:"warningThresholdMs"`
	DNSHostname          *string           `json:"dnsHostname,omitempty"`
	DNSServers           []string          `json:"dnsServers,omitempty"`
	DNSRecords           []DNSRecordConfig `json:"dnsRecords,omitempty"`
	DNSMonitoringMode    *string           `json:"dnsMonitoringMode,omitempty"`
}

func (c *Client) CreateMonitor(input MonitorInput) (*Monitor, error) {
	body, err := c.post("/api/v1/monitor", input)
	if err != nil {
		return nil, err
	}
	var result Monitor
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling monitor: %w", err)
	}
	return &result, nil
}

func (c *Client) GetMonitor(id int) (*Monitor, error) {
	body, err := c.get(fmt.Sprintf("/api/v1/monitor/%d", id))
	if err != nil {
		return nil, err
	}
	var result Monitor
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling monitor: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateMonitor(id int, input MonitorInput) (*Monitor, error) {
	body, err := c.put(fmt.Sprintf("/api/v1/monitor/%d", id), input)
	if err != nil {
		return nil, err
	}
	var result Monitor
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling monitor: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteMonitor(id int) error {
	return c.delete(fmt.Sprintf("/api/v1/monitor/%d", id))
}

func (c *Client) ListMonitors() ([]Monitor, error) {
	return AutoPaginate[Monitor](c, "/api/v1/monitor")
}

func (c *Client) ListMonitorsWithSearch(search string) ([]Monitor, error) {
	return AutoPaginate[Monitor](c, "/api/v1/monitor?value="+url.QueryEscape(search))
}
