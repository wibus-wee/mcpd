package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"mcpv/internal/infra/fsutil"
)

func TestResolveProfilePath(t *testing.T) {
	root := t.TempDir()
	profilesDir := filepath.Join(root, profilesDirName)
	require.NoError(t, os.MkdirAll(profilesDir, fsutil.DefaultDirMode))

	defaultPath := filepath.Join(profilesDir, "default.yaml")
	require.NoError(t, os.WriteFile(defaultPath, []byte("servers: []\n"), fsutil.DefaultFileMode))

	got, err := ResolveProfilePath(root, "default")
	require.NoError(t, err)
	require.Equal(t, defaultPath, got)

	otherPath := filepath.Join(profilesDir, "other.yml")
	require.NoError(t, os.WriteFile(otherPath, []byte("servers: []\n"), fsutil.DefaultFileMode))

	got, err = ResolveProfilePath(root, "other")
	require.NoError(t, err)
	require.Equal(t, otherPath, got)
}

func TestResolveProfilePath_DuplicateExtensions(t *testing.T) {
	root := t.TempDir()
	profilesDir := filepath.Join(root, profilesDirName)
	require.NoError(t, os.MkdirAll(profilesDir, fsutil.DefaultDirMode))

	require.NoError(t, os.WriteFile(filepath.Join(profilesDir, "dup.yaml"), []byte("servers: []\n"), fsutil.DefaultFileMode))
	require.NoError(t, os.WriteFile(filepath.Join(profilesDir, "dup.yml"), []byte("servers: []\n"), fsutil.DefaultFileMode))

	_, err := ResolveProfilePath(root, "dup")
	require.Error(t, err)
}

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
