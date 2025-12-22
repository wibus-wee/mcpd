package catalog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"mcpd/internal/domain"
)

type Loader struct {
	logger *zap.Logger
}

func NewLoader(logger *zap.Logger) *Loader {
	if logger == nil {
		return &Loader{logger: zap.NewNop()}
	}
	return &Loader{logger: logger.Named("catalog")}
}

func (l *Loader) Load(ctx context.Context, path string) (map[string]domain.ServerSpec, error) {
	if path == "" {
		return nil, errors.New("config path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	expanded := os.ExpandEnv(string(data))

	v := viper.New()

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if ext == "" {
		ext = "yaml"
	}
	v.SetConfigType(ext)

	if err := v.ReadConfig(bytes.NewBufferString(expanded)); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	var cfg struct {
		Servers []domain.ServerSpec `mapstructure:"servers"`
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(cfg.Servers) == 0 {
		return nil, errors.New("no servers defined in catalog")
	}

	specs := make(map[string]domain.ServerSpec, len(cfg.Servers))
	var validationErrors []string
	nameSeen := make(map[string]struct{})

	for i, spec := range cfg.Servers {
		if _, exists := nameSeen[spec.Name]; exists {
			validationErrors = append(validationErrors, fmt.Sprintf("servers[%d]: duplicate name %q", i, spec.Name))
		} else if spec.Name != "" {
			nameSeen[spec.Name] = struct{}{}
		}

		if errs := validateServerSpec(spec, i); len(errs) > 0 {
			validationErrors = append(validationErrors, errs...)
			continue
		}

		specs[spec.Name] = spec
	}

	if len(validationErrors) > 0 {
		return nil, errors.New(strings.Join(validationErrors, "; "))
	}

	return specs, nil
}

func validateServerSpec(spec domain.ServerSpec, index int) []string {
	var errs []string

	if spec.Name == "" {
		errs = append(errs, fmt.Sprintf("servers[%d]: name is required", index))
	}
	if len(spec.Cmd) == 0 {
		errs = append(errs, fmt.Sprintf("servers[%d]: cmd is required", index))
	}
	if spec.MaxConcurrent < 1 {
		errs = append(errs, fmt.Sprintf("servers[%d]: maxConcurrent must be >= 1", index))
	}
	if spec.IdleSeconds < 0 {
		errs = append(errs, fmt.Sprintf("servers[%d]: idleSeconds must be >= 0", index))
	}
	if spec.MinReady < 0 {
		errs = append(errs, fmt.Sprintf("servers[%d]: minReady must be >= 0", index))
	}

	versionPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if spec.ProtocolVersion == "" {
		errs = append(errs, fmt.Sprintf("servers[%d]: protocolVersion is required", index))
	} else {
		if !versionPattern.MatchString(spec.ProtocolVersion) {
			errs = append(errs, fmt.Sprintf("servers[%d]: protocolVersion must match YYYY-MM-DD", index))
		}
		if spec.ProtocolVersion != domain.DefaultProtocolVersion {
			errs = append(errs, fmt.Sprintf("servers[%d]: protocolVersion must be %s", index, domain.DefaultProtocolVersion))
		}
	}

	return errs
}
