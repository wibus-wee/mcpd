package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSelector_PathServer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/mcp/server/context7", nil)
	sel, err := ParseSelector(req, "/mcp")
	require.NoError(t, err)
	require.Equal(t, "context7", sel.Server)
	require.Empty(t, sel.Tags)
}

func TestParseSelector_PathTags(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/mcp/tags/git,db", nil)
	sel, err := ParseSelector(req, "/mcp")
	require.NoError(t, err)
	require.Empty(t, sel.Server)
	require.Equal(t, []string{"db", "git"}, sel.Tags)
}

func TestParseSelector_HeaderServer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/mcp", nil)
	req.Header.Set("X-Mcp-Server", "context7")
	sel, err := ParseSelector(req, "/mcp")
	require.NoError(t, err)
	require.Equal(t, "context7", sel.Server)
}

func TestParseSelector_HeaderTags(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/mcp", nil)
	req.Header.Set("X-Mcp-Tags", "Git, db")
	sel, err := ParseSelector(req, "/mcp")
	require.NoError(t, err)
	require.Equal(t, []string{"db", "git"}, sel.Tags)
}

func TestParseSelector_PathAndHeaderConflict(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/mcp/server/context7", nil)
	req.Header.Set("X-Mcp-Server", "other")
	_, err := ParseSelector(req, "/mcp")
	require.Error(t, err)
}

func TestParseSelector_MissingSelector(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/mcp", nil)
	_, err := ParseSelector(req, "/mcp")
	require.ErrorIs(t, err, ErrSelectorRequired)
}
