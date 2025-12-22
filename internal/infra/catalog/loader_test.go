package catalog

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"mcpd/internal/domain"
)

func TestLoader_Success(t *testing.T) {
	file := writeTempConfig(t, `
servers:
  - name: git-helper
    cmd: ["./git-helper"]
    idleSeconds: 60
    maxConcurrent: 2
    sticky: false
    persistent: false
    minReady: 0
    protocolVersion: "2025-11-25"
`)

	loader := NewLoader(zap.NewNop())
	specs, err := loader.Load(context.Background(), file)
	require.NoError(t, err)
	require.Len(t, specs, 1)

	got := specs["git-helper"]
	expect := domain.ServerSpec{
		Name:            "git-helper",
		Cmd:             []string{"./git-helper"},
		IdleSeconds:     60,
		MaxConcurrent:   2,
		Sticky:          false,
		Persistent:      false,
		MinReady:        0,
		ProtocolVersion: domain.DefaultProtocolVersion,
	}
	if diff := cmp.Diff(expect, got); diff != "" {
		t.Fatalf("spec mismatch (-want +got):\n%s", diff)
	}
}

func TestLoader_EnvExpansion(t *testing.T) {
	t.Setenv("SERVER_CMD", "./from-env")
	file := writeTempConfig(t, `
servers:
  - name: env-server
    cmd: ["${SERVER_CMD}"]
    idleSeconds: 0
    maxConcurrent: 1
    sticky: false
    persistent: false
    minReady: 0
    protocolVersion: "2025-11-25"
`)

	loader := NewLoader(zap.NewNop())
	specs, err := loader.Load(context.Background(), file)
	require.NoError(t, err)
	require.Equal(t, []string{"./from-env"}, specs["env-server"].Cmd)
}

func TestLoader_DuplicateName(t *testing.T) {
	file := writeTempConfig(t, `
servers:
  - name: dup
    cmd: ["./a"]
    idleSeconds: 0
    maxConcurrent: 1
    sticky: false
    persistent: false
    minReady: 0
    protocolVersion: "2025-11-25"
  - name: dup
    cmd: ["./b"]
    idleSeconds: 0
    maxConcurrent: 1
    sticky: false
    persistent: false
    minReady: 0
    protocolVersion: "2025-11-25"
`)

	loader := NewLoader(zap.NewNop())
	_, err := loader.Load(context.Background(), file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate name")
}

func TestLoader_InvalidProtocolVersion(t *testing.T) {
	file := writeTempConfig(t, `
servers:
  - name: bad-protocol
    cmd: ["./a"]
    idleSeconds: 0
    maxConcurrent: 1
    sticky: false
    persistent: false
    minReady: 0
    protocolVersion: "2024-01"
`)

	loader := NewLoader(zap.NewNop())
	_, err := loader.Load(context.Background(), file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "protocolVersion must match")
}

func TestLoader_MissingRequiredFields(t *testing.T) {
	file := writeTempConfig(t, `
servers:
  - name: ""
    cmd: []
    idleSeconds: -1
    maxConcurrent: 0
    sticky: false
    persistent: false
    minReady: -2
    protocolVersion: ""
`)

	loader := NewLoader(zap.NewNop())
	_, err := loader.Load(context.Background(), file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "name is required")
	require.Contains(t, err.Error(), "cmd is required")
	require.Contains(t, err.Error(), "maxConcurrent must be")
	require.Contains(t, err.Error(), "idleSeconds must be")
	require.Contains(t, err.Error(), "minReady must be")
	require.Contains(t, err.Error(), "protocolVersion is required")
}

func TestLoader_NoServers(t *testing.T) {
	file := writeTempConfig(t, `
servers: []
`)

	loader := NewLoader(zap.NewNop())
	_, err := loader.Load(context.Background(), file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no servers defined")
}

func TestLoader_ContextCanceled(t *testing.T) {
	file := writeTempConfig(t, `
servers:
  - name: ok
    cmd: ["./a"]
    idleSeconds: 0
    maxConcurrent: 1
    sticky: false
    persistent: false
    minReady: 0
    protocolVersion: "2025-11-25"
`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	loader := NewLoader(zap.NewNop())
	_, err := loader.Load(ctx, file)
	require.ErrorIs(t, err, context.Canceled)
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}
