package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func ResolveProfilePath(storePath string, profileName string) (string, error) {
	if storePath == "" {
		return "", errors.New("profile store path is required")
	}
	if profileName == "" {
		return "", errors.New("profile name is required")
	}

	profilesDir := filepath.Join(storePath, profilesDirName)
	candidateYAML := filepath.Join(profilesDir, profileName+".yaml")
	candidateYML := filepath.Join(profilesDir, profileName+".yml")

	yamlExists, err := fileExists(candidateYAML)
	if err != nil {
		return "", err
	}
	ymlExists, err := fileExists(candidateYML)
	if err != nil {
		return "", err
	}

	if yamlExists && ymlExists {
		return "", fmt.Errorf("profile %q has both .yaml and .yml files", profileName)
	}
	if yamlExists {
		return candidateYAML, nil
	}
	if ymlExists {
		return candidateYML, nil
	}

	return "", fmt.Errorf("profile %q not found in %s", profileName, profilesDir)
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

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat %s: %w", path, err)
}
