package loader

import (
	"bytes"
	"fmt"

	"mcpv/internal/infra/catalog/normalizer"
)

func decodeRuntimeConfig(expanded string) (normalizer.RawRuntimeConfig, error) {
	v := newRuntimeViper()
	if err := v.ReadConfig(bytes.NewBufferString(expanded)); err != nil {
		return normalizer.RawRuntimeConfig{}, fmt.Errorf("parse config: %w", err)
	}
	var cfg normalizer.RawRuntimeConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return normalizer.RawRuntimeConfig{}, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

func decodeCatalog(expanded string) (normalizer.RawCatalog, error) {
	v := newRuntimeViper()
	if err := v.ReadConfig(bytes.NewBufferString(expanded)); err != nil {
		return normalizer.RawCatalog{}, fmt.Errorf("parse config: %w", err)
	}
	var cfg normalizer.RawCatalog
	if err := v.Unmarshal(&cfg); err != nil {
		return normalizer.RawCatalog{}, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}
