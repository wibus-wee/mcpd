package ui

import (
	"encoding/json"
	"strings"
)

const GatewaySectionKey = "gateway"

type GatewaySettings struct {
	Enabled    bool
	BinaryPath string
	Args       []string
	HTTPAddr   string
	HTTPPath   string
	HTTPToken  string
	Caller     string
	RPC        string
	HealthURL  string
}

type gatewaySettingsPayload struct {
	Enabled    *bool    `json:"enabled,omitempty"`
	BinaryPath string   `json:"binaryPath,omitempty"`
	Args       []string `json:"args,omitempty"`
	HTTPAddr   string   `json:"httpAddr,omitempty"`
	HTTPPath   string   `json:"httpPath,omitempty"`
	HTTPToken  string   `json:"httpToken,omitempty"`
	Caller     string   `json:"caller,omitempty"`
	RPC        string   `json:"rpc,omitempty"`
	HealthURL  string   `json:"healthUrl,omitempty"`
}

func DefaultGatewaySettings() GatewaySettings {
	return GatewaySettings{
		Enabled:   true,
		HTTPAddr:  "127.0.0.1:8090",
		HTTPPath:  "/mcp",
		Caller:    "mcpvmcp-ui",
		HealthURL: "",
	}
}

func ParseGatewaySettings(raw json.RawMessage) (GatewaySettings, error) {
	settings := DefaultGatewaySettings()
	if len(raw) == 0 {
		return settings, nil
	}
	var payload gatewaySettingsPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return settings, err
	}
	if payload.Enabled != nil {
		settings.Enabled = *payload.Enabled
	}
	if payload.BinaryPath != "" {
		settings.BinaryPath = payload.BinaryPath
	}
	if payload.Args != nil {
		settings.Args = append([]string(nil), payload.Args...)
	}
	if payload.HTTPAddr != "" {
		settings.HTTPAddr = payload.HTTPAddr
	}
	if payload.HTTPPath != "" {
		settings.HTTPPath = payload.HTTPPath
	}
	if payload.HTTPToken != "" {
		settings.HTTPToken = payload.HTTPToken
	}
	if payload.Caller != "" {
		settings.Caller = payload.Caller
	}
	if payload.RPC != "" {
		settings.RPC = payload.RPC
	}
	if payload.HealthURL != "" {
		settings.HealthURL = payload.HealthURL
	}
	return settings, nil
}

func BuildGatewayProcessConfig(settings GatewaySettings) GatewayProcessConfig {
	cfg := GatewayProcessConfig{
		Enabled:    settings.Enabled,
		BinaryPath: strings.TrimSpace(settings.BinaryPath),
		HealthURL:  strings.TrimSpace(settings.HealthURL),
	}
	if cfg.BinaryPath == "" {
		cfg.BinaryPath = ResolveMcpvmcpPath()
	}
	if len(settings.Args) > 0 {
		cfg.Args = append([]string(nil), settings.Args...)
		return cfg
	}

	httpAddr := strings.TrimSpace(settings.HTTPAddr)
	if httpAddr == "" {
		httpAddr = "127.0.0.1:8090"
	}
	httpPath := strings.TrimSpace(settings.HTTPPath)
	if httpPath == "" {
		httpPath = "/mcp"
	}
	if !strings.HasPrefix(httpPath, "/") {
		httpPath = "/" + httpPath
	}
	caller := strings.TrimSpace(settings.Caller)
	if caller == "" {
		caller = "mcpvmcp-ui"
	}

	args := []string{
		"--transport", "streamable-http",
		"--http-addr", httpAddr,
		"--http-path", httpPath,
		"--caller", caller,
	}
	if rpcAddr := strings.TrimSpace(settings.RPC); rpcAddr != "" {
		args = append(args, "--rpc", rpcAddr)
	}
	if token := strings.TrimSpace(settings.HTTPToken); token != "" {
		args = append(args, "--http-token", token)
	}
	cfg.Args = args

	if cfg.HealthURL == "" {
		base := httpAddr
		if !strings.Contains(base, "://") {
			base = "http://" + base
		}
		cfg.HealthURL = strings.TrimRight(base, "/") + "/healthz"
	}
	return cfg
}
