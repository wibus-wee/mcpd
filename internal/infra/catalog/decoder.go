package catalog

import (
	"bytes"
	"fmt"
)

func decodeRuntimeConfig(expanded string) (rawRuntimeConfig, error) {
	v := newRuntimeViper()
	if err := v.ReadConfig(bytes.NewBufferString(expanded)); err != nil {
		return rawRuntimeConfig{}, fmt.Errorf("parse config: %w", err)
	}
	var cfg rawRuntimeConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return rawRuntimeConfig{}, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

func decodeCatalog(expanded string) (rawCatalog, error) {
	v := newRuntimeViper()
	if err := v.ReadConfig(bytes.NewBufferString(expanded)); err != nil {
		return rawCatalog{}, fmt.Errorf("parse config: %w", err)
	}
	var cfg rawCatalog
	if err := v.Unmarshal(&cfg); err != nil {
		return rawCatalog{}, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}
