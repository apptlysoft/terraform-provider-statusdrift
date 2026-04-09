package notification_channel

import "testing"

func TestIsMaskedValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "full mask asterisks",
			input:    "****",
			expected: true,
		},
		{
			name:     "partial mask webhook URL",
			input:    "http****xxxx",
			expected: true,
		},
		{
			name:     "partial mask bot token",
			input:    "1234****wxyz",
			expected: true,
		},
		{
			name:     "partial mask auth header",
			input:    "Bear****c123",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "real webhook URL",
			input:    "https://hooks.slack.com/services/T00/B00/xxxx",
			expected: false,
		},
		{
			name:     "normal value",
			input:    "normal_value",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMaskedValue(tt.input)
			if got != tt.expected {
				t.Errorf("isMaskedValue(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"webhook_url", "webhookUrl"},
		{"bot_token", "botToken"},
		{"auth_header", "authHeader"},
		{"rest_endpoint_url", "restEndpointUrl"},
		{"api_key", "apiKey"},
		{"email", "email"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := snakeToCamel(tt.input)
			if got != tt.expected {
				t.Errorf("snakeToCamel(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"webhookUrl", "webhook_url"},
		{"botToken", "bot_token"},
		{"authHeader", "auth_header"},
		{"restEndpointUrl", "rest_endpoint_url"},
		{"apiKey", "api_key"},
		{"email", "email"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := camelToSnake(tt.input)
			if got != tt.expected {
				t.Errorf("camelToSnake(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
