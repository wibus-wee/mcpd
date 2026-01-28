package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDeepLink_ValidURLs(t *testing.T) {
	tests := []struct {
		name       string
		rawURL     string
		wantPath   string
		wantParams map[string]string
	}{
		{
			name:       "simple path",
			rawURL:     "mcpd://servers",
			wantPath:   "/servers",
			wantParams: map[string]string{},
		},
		{
			name:       "path with query params",
			rawURL:     "mcpd://servers?tab=overview",
			wantPath:   "/servers",
			wantParams: map[string]string{"tab": "overview"},
		},
		{
			name:       "nested path",
			rawURL:     "mcpd://servers/detail",
			wantPath:   "/servers/detail",
			wantParams: map[string]string{},
		},
		{
			name:       "nested path with params",
			rawURL:     "mcpd://servers/detail?server=test&tab=config",
			wantPath:   "/servers/detail",
			wantParams: map[string]string{"server": "test", "tab": "config"},
		},
		{
			name:       "root path",
			rawURL:     "mcpd://",
			wantPath:   "/",
			wantParams: map[string]string{},
		},
		{
			name:       "settings",
			rawURL:     "mcpd://settings",
			wantPath:   "/settings",
			wantParams: map[string]string{},
		},
		{
			name:       "dev scheme simple path",
			rawURL:     "mcpdev://servers",
			wantPath:   "/servers",
			wantParams: map[string]string{},
		},
		{
			name:       "dev scheme with params",
			rawURL:     "mcpdev://servers?tab=overview",
			wantPath:   "/servers",
			wantParams: map[string]string{"tab": "overview"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link, err := ParseDeepLink(tt.rawURL)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, link.Path())
			assert.Equal(t, tt.wantParams, link.Params())
			assert.Equal(t, tt.rawURL, link.Raw())
		})
	}
}

func TestParseDeepLink_InvalidURLs(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
	}{
		{
			name:   "empty URL",
			rawURL: "",
		},
		{
			name:   "wrong scheme",
			rawURL: "https://example.com",
		},
		{
			name:   "wrong scheme http",
			rawURL: "http://servers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link, err := ParseDeepLink(tt.rawURL)
			assert.Error(t, err)
			assert.Nil(t, link)
		})
	}
}
