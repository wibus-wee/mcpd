package rpc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"mcpv/internal/domain"
)

const (
	authorizationHeader = "authorization"
	bearerPrefix        = "bearer "
)

type resolvedAuth struct {
	enabled bool
	mode    domain.RPCAuthMode
	token   string
}

func resolveServerAuth(cfg domain.RPCAuthConfig) (resolvedAuth, error) {
	if !cfg.Enabled {
		return resolvedAuth{enabled: false}, nil
	}
	mode := cfg.Mode
	if mode == "" {
		mode = domain.RPCAuthModeToken
	}
	switch mode {
	case domain.RPCAuthModeToken:
		token, err := resolveToken(cfg)
		if err != nil {
			return resolvedAuth{}, err
		}
		return resolvedAuth{
			enabled: true,
			mode:    mode,
			token:   token,
		}, nil
	case domain.RPCAuthModeMTLS:
		return resolvedAuth{
			enabled: true,
			mode:    mode,
		}, nil
	default:
		return resolvedAuth{}, fmt.Errorf("unsupported rpc auth mode: %s", mode)
	}
}

func resolveClientToken(cfg domain.RPCAuthConfig) (string, error) {
	if cfg.Mode != "" && cfg.Mode != domain.RPCAuthModeToken {
		return "", nil
	}
	if !cfg.Enabled && strings.TrimSpace(cfg.Token) == "" && strings.TrimSpace(cfg.TokenEnv) == "" {
		return "", nil
	}
	return resolveToken(cfg)
}

func resolveToken(cfg domain.RPCAuthConfig) (string, error) {
	if token := strings.TrimSpace(cfg.Token); token != "" {
		return token, nil
	}
	envKey := strings.TrimSpace(cfg.TokenEnv)
	if envKey == "" {
		return "", domain.E(domain.CodeInvalidArgument, "rpc.auth", "rpc auth token is required", nil)
	}
	value, ok := os.LookupEnv(envKey)
	if !ok || strings.TrimSpace(value) == "" {
		return "", domain.E(domain.CodeInvalidArgument, "rpc.auth", fmt.Sprintf("rpc auth token env %q is not set", envKey), nil)
	}
	return strings.TrimSpace(value), nil
}

func authorizeContext(ctx context.Context, auth resolvedAuth) error {
	if !auth.enabled {
		return nil
	}
	switch auth.mode {
	case domain.RPCAuthModeToken:
		return authorizeToken(ctx, auth.token)
	case domain.RPCAuthModeMTLS:
		return authorizeMTLS(ctx)
	default:
		return status.Error(codes.Unauthenticated, "invalid auth mode")
	}
}

func authorizeToken(ctx context.Context, expected string) error {
	if strings.TrimSpace(expected) == "" {
		return status.Error(codes.Unauthenticated, "authorization required")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "authorization required")
	}
	values := md.Get(authorizationHeader)
	for _, value := range values {
		token := parseBearerToken(value)
		if token != "" && token == expected {
			return nil
		}
	}
	return status.Error(codes.Unauthenticated, "invalid authorization token")
}

func authorizeMTLS(ctx context.Context) error {
	info, ok := peer.FromContext(ctx)
	if !ok || info == nil || info.AuthInfo == nil {
		return status.Error(codes.Unauthenticated, "client certificate required")
	}
	tlsInfo, ok := info.AuthInfo.(credentials.TLSInfo)
	if !ok || len(tlsInfo.State.VerifiedChains) == 0 {
		return status.Error(codes.Unauthenticated, "client certificate required")
	}
	return nil
}

func parseBearerToken(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, bearerPrefix) {
		return strings.TrimSpace(trimmed[len(bearerPrefix):])
	}
	return ""
}

func attachBearerToken(ctx context.Context, token string) context.Context {
	if ctx == nil || strings.TrimSpace(token) == "" {
		return ctx
	}
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.New(nil)
	}
	if len(md.Get(authorizationHeader)) == 0 {
		md.Set(authorizationHeader, fmt.Sprintf("Bearer %s", token))
	}
	return metadata.NewOutgoingContext(ctx, md)
}
