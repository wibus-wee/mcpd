package normalizer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRPCAuthConfig(t *testing.T) {
	base := RawRPCConfig{
		ListenAddress:           "unix:///tmp/mcpv.sock",
		MaxRecvMsgSize:          1024,
		MaxSendMsgSize:          1024,
		KeepaliveTimeSeconds:    0,
		KeepaliveTimeoutSeconds: 0,
		SocketMode:              "0660",
		TLS: RawRPCTLSConfig{
			Enabled:    false,
			CertFile:   "",
			KeyFile:    "",
			CAFile:     "",
			ClientAuth: false,
		},
	}

	tests := []struct {
		name     string
		cfg      RawRPCConfig
		wantErrs []string
	}{
		{
			name: "token missing",
			cfg: func() RawRPCConfig {
				cfg := base
				cfg.Auth = RawRPCAuthConfig{Enabled: true, Mode: "token"}
				return cfg
			}(),
			wantErrs: []string{"rpc.auth.token or rpc.auth.tokenEnv"},
		},
		{
			name: "token tcp requires tls",
			cfg: func() RawRPCConfig {
				cfg := base
				cfg.ListenAddress = "tcp://127.0.0.1:7011"
				cfg.Auth = RawRPCAuthConfig{Enabled: true, Mode: "token", Token: "secret"}
				return cfg
			}(),
			wantErrs: []string{"rpc.tls.enabled is required"},
		},
		{
			name: "token unix ok",
			cfg: func() RawRPCConfig {
				cfg := base
				cfg.Auth = RawRPCAuthConfig{Enabled: true, Mode: "token", Token: "secret"}
				return cfg
			}(),
			wantErrs: nil,
		},
		{
			name: "mtls requires tls",
			cfg: func() RawRPCConfig {
				cfg := base
				cfg.ListenAddress = "tcp://127.0.0.1:7012"
				cfg.Auth = RawRPCAuthConfig{Enabled: true, Mode: "mtls"}
				return cfg
			}(),
			wantErrs: []string{"rpc.tls.enabled is required"},
		},
		{
			name: "mtls requires client auth",
			cfg: func() RawRPCConfig {
				cfg := base
				cfg.ListenAddress = "tcp://127.0.0.1:7013"
				cfg.TLS = RawRPCTLSConfig{Enabled: true}
				cfg.Auth = RawRPCAuthConfig{Enabled: true, Mode: "mtls"}
				return cfg
			}(),
			wantErrs: []string{"rpc.tls.clientAuth must be true"},
		},
		{
			name: "mtls ok",
			cfg: func() RawRPCConfig {
				cfg := base
				cfg.ListenAddress = "tcp://127.0.0.1:7014"
				cfg.TLS = RawRPCTLSConfig{Enabled: true, ClientAuth: true, CAFile: "/tmp/ca.pem", CertFile: "/tmp/cert.pem", KeyFile: "/tmp/key.pem"}
				cfg.Auth = RawRPCAuthConfig{Enabled: true, Mode: "mtls"}
				return cfg
			}(),
			wantErrs: nil,
		},
		{
			name: "token and tokenEnv conflict",
			cfg: func() RawRPCConfig {
				cfg := base
				cfg.Auth = RawRPCAuthConfig{Enabled: true, Mode: "token", Token: "secret", TokenEnv: "MCPV_RPC_TOKEN"}
				return cfg
			}(),
			wantErrs: []string{"rpc.auth.token and rpc.auth.tokenEnv are mutually exclusive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := normalizeRPCConfig(tt.cfg)
			if len(tt.wantErrs) == 0 {
				require.Empty(t, errs)
				return
			}
			require.NotEmpty(t, errs)
			combined := strings.Join(errs, "\n")
			for _, want := range tt.wantErrs {
				require.Contains(t, combined, want)
			}
		})
	}
}
