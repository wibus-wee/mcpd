package catalog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"mcpd/internal/infra/fsutil"
)

type RuntimeUpdate struct {
	Path string
	Data []byte
}

type RuntimeConfigUpdate struct {
	RouteTimeoutSeconds        int
	PingIntervalSeconds        int
	ToolRefreshSeconds         int
	ToolRefreshConcurrency     int
	CallerCheckSeconds         int
	CallerInactiveSeconds      int
	ServerInitRetryBaseSeconds int
	ServerInitRetryMaxSeconds  int
	ServerInitMaxRetries       int
	BootstrapMode              string
	BootstrapConcurrency       int
	BootstrapTimeoutSeconds    int
	DefaultActivationMode      string
	ExposeTools                bool
	ToolNamespaceStrategy      string
}

func ResolveRuntimePath(storePath string, allowCreate bool) (string, error) {
	if storePath == "" {
		return "", errors.New("profile store path is required")
	}

	runtimePath := filepath.Join(storePath, runtimeFileName)
	altPath := filepath.Join(storePath, runtimeFileAlt)

	yamlExists, err := fileExists(runtimePath)
	if err != nil {
		return "", err
	}
	ymlExists, err := fileExists(altPath)
	if err != nil {
		return "", err
	}

	if yamlExists && ymlExists {
		return "", fmt.Errorf("runtime config has both %s and %s", runtimeFileName, runtimeFileAlt)
	}
	if yamlExists {
		return runtimePath, nil
	}
	if ymlExists {
		return altPath, nil
	}
	if allowCreate {
		return runtimePath, nil
	}
	return "", fmt.Errorf("runtime config not found in %s", storePath)
}

func UpdateRuntimeConfig(path string, update RuntimeConfigUpdate) (RuntimeUpdate, error) {
	if path == "" {
		return RuntimeUpdate{}, errors.New("runtime config path is required")
	}

	doc, err := loadRuntimeDocument(path)
	if err != nil {
		return RuntimeUpdate{}, err
	}

	doc["routeTimeoutSeconds"] = update.RouteTimeoutSeconds
	doc["pingIntervalSeconds"] = update.PingIntervalSeconds
	doc["toolRefreshSeconds"] = update.ToolRefreshSeconds
	doc["toolRefreshConcurrency"] = update.ToolRefreshConcurrency
	doc["callerCheckSeconds"] = update.CallerCheckSeconds
	doc["callerInactiveSeconds"] = update.CallerInactiveSeconds
	doc["serverInitRetryBaseSeconds"] = update.ServerInitRetryBaseSeconds
	doc["serverInitRetryMaxSeconds"] = update.ServerInitRetryMaxSeconds
	doc["serverInitMaxRetries"] = update.ServerInitMaxRetries
	doc["bootstrapMode"] = strings.TrimSpace(update.BootstrapMode)
	doc["bootstrapConcurrency"] = update.BootstrapConcurrency
	doc["bootstrapTimeoutSeconds"] = update.BootstrapTimeoutSeconds
	doc["defaultActivationMode"] = strings.TrimSpace(update.DefaultActivationMode)
	doc["exposeTools"] = update.ExposeTools
	doc["toolNamespaceStrategy"] = strings.TrimSpace(update.ToolNamespaceStrategy)

	merged, err := yaml.Marshal(doc)
	if err != nil {
		return RuntimeUpdate{}, fmt.Errorf("render runtime config: %w", err)
	}

	return RuntimeUpdate{Path: path, Data: merged}, nil
}

func loadRuntimeDocument(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("read runtime config: %w", err)
	}

	doc := make(map[string]any)
	if len(data) == 0 {
		return doc, nil
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse runtime config: %w", err)
	}
	return doc, nil
}

func writeRuntimeUpdate(update RuntimeUpdate) error {
	if err := os.MkdirAll(filepath.Dir(update.Path), fsutil.DefaultDirMode); err != nil {
		return fmt.Errorf("create runtime config directory: %w", err)
	}
	if err := os.WriteFile(update.Path, update.Data, fsutil.DefaultFileMode); err != nil {
		return fmt.Errorf("write runtime config: %w", err)
	}
	return nil
}
