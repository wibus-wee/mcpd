package editor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"mcpv/internal/infra/fsutil"
)

func TestUpdateRuntimeConfig_PreservesOtherFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.yaml")
	content := `routeTimeoutSeconds: 10
subAgent:
  model: "gpt-4o"
rpc:
  listenAddress: "unix:///tmp/test.sock"
`
	require.NoError(t, os.WriteFile(path, []byte(content), fsutil.DefaultFileMode))

	update, err := UpdateRuntimeConfig(path, RuntimeConfigUpdate{
		RouteTimeoutSeconds:        15,
		PingIntervalSeconds:        20,
		ToolRefreshSeconds:         30,
		ToolRefreshConcurrency:     4,
		ClientCheckSeconds:         5,
		ClientInactiveSeconds:      60,
		ServerInitRetryBaseSeconds: 1,
		ServerInitRetryMaxSeconds:  5,
		ServerInitMaxRetries:       2,
		ReloadMode:                 "strict",
		BootstrapMode:              "metadata",
		BootstrapConcurrency:       3,
		BootstrapTimeoutSeconds:    15,
		DefaultActivationMode:      "on-demand",
		ExposeTools:                true,
		ToolNamespaceStrategy:      "flat",
	})
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(update.Data, &doc))
	require.Equal(t, 15, doc["routeTimeoutSeconds"])
	require.Equal(t, "strict", doc["reloadMode"])

	subAgent, ok := doc["subAgent"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "gpt-4o", subAgent["model"])

	rpc, ok := doc["rpc"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "unix:///tmp/test.sock", rpc["listenAddress"])
}

func TestUpdateSubAgentConfig_PreservesOtherFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.yaml")
	content := `routeTimeoutSeconds: 10
subAgent:
  model: "gpt-4o"
  apiKey: "secret"
  baseURL: "https://example.com"
rpc:
  listenAddress: "unix:///tmp/test.sock"
`
	require.NoError(t, os.WriteFile(path, []byte(content), fsutil.DefaultFileMode))

	model := "gpt-4.1"
	provider := "openai"
	apiKeyEnvVar := "OPENAI_API_KEY"
	baseURL := ""
	maxTools := 12
	filterPrompt := "filter prompt"

	update, err := UpdateSubAgentConfig(path, SubAgentConfigUpdate{
		Model:              &model,
		Provider:           &provider,
		APIKeyEnvVar:       &apiKeyEnvVar,
		BaseURL:            &baseURL,
		MaxToolsPerRequest: &maxTools,
		FilterPrompt:       &filterPrompt,
	})
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(update.Data, &doc))
	require.Equal(t, 10, doc["routeTimeoutSeconds"])

	subAgent, ok := doc["subAgent"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "gpt-4.1", subAgent["model"])
	require.Equal(t, "openai", subAgent["provider"])
	require.Equal(t, "OPENAI_API_KEY", subAgent["apiKeyEnvVar"])
	require.Equal(t, "", subAgent["baseURL"])
	require.Equal(t, 12, subAgent["maxToolsPerRequest"])
	require.Equal(t, "filter prompt", subAgent["filterPrompt"])
	require.Equal(t, "secret", subAgent["apiKey"])

	rpc, ok := doc["rpc"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "unix:///tmp/test.sock", rpc["listenAddress"])
}
