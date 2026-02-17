package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func loadJSONPayload(arg string, filePath string) ([]byte, error) {
	if arg != "" && filePath != "" {
		return nil, errors.New("--args and --args-file are mutually exclusive")
	}
	var payload []byte
	switch {
	case filePath != "":
		data, err := os.ReadFile(filepath.Clean(filePath))
		if err != nil {
			return nil, fmt.Errorf("read args file: %w", err)
		}
		payload = data
	case arg != "":
		payload = []byte(arg)
	default:
		payload = []byte("{}")
	}
	if !json.Valid(payload) {
		return nil, errors.New("invalid JSON payload")
	}
	return payload, nil
}
