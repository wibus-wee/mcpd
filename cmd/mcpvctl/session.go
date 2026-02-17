package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mcpv/internal/domain"
	"mcpv/internal/infra/rpc"
	controlv1 "mcpv/pkg/api/control/v1"
)

func withClient(ctx context.Context, opts *cliOptions, fn func(context.Context, controlv1.ControlPlaneServiceClient) error) error {
	client, err := dialClient(ctx, opts)
	if err != nil {
		return err
	}
	defer client.Close()
	return fn(ctx, client.Control())
}

func withSession(ctx context.Context, opts *cliOptions, fn func(context.Context, controlv1.ControlPlaneServiceClient, string) error) error {
	client, err := dialClient(ctx, opts)
	if err != nil {
		return err
	}
	defer client.Close()

	caller := resolveCaller(opts.caller)
	if !opts.noRegister {
		if err := validateSelectorFlags(opts); err != nil {
			return err
		}
		_, err := client.Control().RegisterCaller(ctx, &controlv1.RegisterCallerRequest{
			Caller: caller,
			Pid:    int64(os.Getpid()),
			Tags:   normalizeTags(opts.tags),
			Server: strings.TrimSpace(opts.server),
		})
		if err != nil {
			return err
		}
		defer func() {
			unregisterCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = client.Control().UnregisterCaller(unregisterCtx, &controlv1.UnregisterCallerRequest{Caller: caller})
		}()
	}

	return fn(ctx, client.Control(), caller)
}

func dialClient(ctx context.Context, opts *cliOptions) (*rpc.Client, error) {
	cfg := rpc.ClientConfig{
		Address:                 opts.rpcAddress,
		MaxRecvMsgSize:          opts.rpcMaxRecvMsgSize,
		MaxSendMsgSize:          opts.rpcMaxSendMsgSize,
		KeepaliveTimeSeconds:    opts.rpcKeepaliveTime,
		KeepaliveTimeoutSeconds: opts.rpcKeepaliveTimeout,
		TLS: domain.RPCTLSConfig{
			Enabled:  opts.rpcTLSEnabled,
			CertFile: strings.TrimSpace(opts.rpcTLSCertFile),
			KeyFile:  strings.TrimSpace(opts.rpcTLSKeyFile),
			CAFile:   strings.TrimSpace(opts.rpcTLSCAFile),
		},
		Auth: domain.RPCAuthConfig{
			Enabled:  strings.TrimSpace(opts.rpcToken) != "" || strings.TrimSpace(opts.rpcTokenEnv) != "",
			Mode:     domain.RPCAuthModeToken,
			Token:    strings.TrimSpace(opts.rpcToken),
			TokenEnv: strings.TrimSpace(opts.rpcTokenEnv),
		},
	}
	return rpc.Dial(ctx, cfg)
}

func resolveCaller(value string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	base := "mcpvctl"
	if pid := os.Getpid(); pid > 0 {
		return fmt.Sprintf("%s-%d", base, pid)
	}
	return base
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func signalAwareContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(signals)
		select {
		case <-signals:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}
