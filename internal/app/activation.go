package app

import "mcpd/internal/domain"

func resolveActivationMode(runtime domain.RuntimeConfig, spec domain.ServerSpec) domain.ActivationMode {
	mode := spec.ActivationMode
	if mode == "" {
		mode = runtime.DefaultActivationMode
	}
	if mode == "" {
		mode = domain.DefaultActivationMode
	}
	return mode
}

func activeMinReady(spec domain.ServerSpec) int {
	if spec.MinReady < 1 {
		return 1
	}
	return spec.MinReady
}

func baselineMinReady(runtime domain.RuntimeConfig, spec domain.ServerSpec) int {
	if resolveActivationMode(runtime, spec) != domain.ActivationAlwaysOn {
		return 0
	}
	return activeMinReady(spec)
}
