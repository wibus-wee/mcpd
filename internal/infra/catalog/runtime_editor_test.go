package catalog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"mcpd/internal/infra/fsutil"
)

func TestResolveRuntimePath(t *testing.T) {
	root := t.TempDir()

	path, err := ResolveRuntimePath(root, true)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(root, runtimeFileName), path)

	altPath := filepath.Join(root, runtimeFileAlt)
	require.NoError(t, os.WriteFile(altPath, []byte("routeTimeoutSeconds: 10\n"), fsutil.DefaultFileMode))

	path, err = ResolveRuntimePath(root, false)
	require.NoError(t, err)
	require.Equal(t, altPath, path)
}

func TestResolveRuntimePath_DuplicateExtensions(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, runtimeFileName), []byte("routeTimeoutSeconds: 10\n"), fsutil.DefaultFileMode))
	require.NoError(t, os.WriteFile(filepath.Join(root, runtimeFileAlt), []byte("routeTimeoutSeconds: 10\n"), fsutil.DefaultFileMode))

	_, err := ResolveRuntimePath(root, false)
	require.Error(t, err)
}

func TestUpdateRuntimeConfig_PreservesOtherFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), runtimeFileName)
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
		CallerCheckSeconds:         5,
		CallerInactiveSeconds:      60,
		ServerInitRetryBaseSeconds: 1,
		ServerInitRetryMaxSeconds:  5,
		ServerInitMaxRetries:       2,
		StartupStrategy:            "eager",
		BootstrapConcurrency:       3,
		BootstrapTimeoutSeconds:    15,
		ExposeTools:                true,
		ToolNamespaceStrategy:      "flat",
	})
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(update.Data, &doc))
	require.Equal(t, 15, doc["routeTimeoutSeconds"])

	subAgent, ok := doc["subAgent"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "gpt-4o", subAgent["model"])

	rpc, ok := doc["rpc"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "unix:///tmp/test.sock", rpc["listenAddress"])
}
