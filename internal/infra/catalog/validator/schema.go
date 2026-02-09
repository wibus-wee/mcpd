package validator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed schema.json
var catalogSchema []byte

var (
	catalogSchemaOnce     sync.Once
	catalogSchemaResolved *jsonschema.Resolved
	catalogSchemaErr      error
)

func ValidateCatalogSchema(raw string) error {
	resolved, err := loadCatalogSchema()
	if err != nil {
		return err
	}

	var payload any
	if err := yaml.Unmarshal([]byte(raw), &payload); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	if err := resolved.Validate(payload); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

func loadCatalogSchema() (*jsonschema.Resolved, error) {
	catalogSchemaOnce.Do(func() {
		var schema jsonschema.Schema
		if err := json.Unmarshal(catalogSchema, &schema); err != nil {
			catalogSchemaErr = fmt.Errorf("parse catalog schema: %w", err)
			return
		}
		resolved, err := schema.Resolve(nil)
		if err != nil {
			catalogSchemaErr = fmt.Errorf("resolve catalog schema: %w", err)
			return
		}
		catalogSchemaResolved = resolved
	})

	if catalogSchemaErr != nil {
		return nil, catalogSchemaErr
	}

	return catalogSchemaResolved, nil
}
