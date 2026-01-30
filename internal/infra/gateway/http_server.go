package gateway

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

const (
	defaultHTTPPath              = "/mcp"
	defaultHTTPReadHeaderTimeout = 10 * time.Second
	defaultHTTPReadTimeout       = 15 * time.Second
	defaultHTTPIdleTimeout       = 60 * time.Second
	defaultHTTPShutdownTimeout   = 5 * time.Second
)

type HTTPOptions struct {
	Addr               string
	Path               string
	Token              string
	AllowedOrigins     []string
	JSONResponse       bool
	SessionTimeout     time.Duration
	TLSEnabled         bool
	TLSCertFile        string
	TLSKeyFile         string
	EventStoreEnabled  bool
	EventStoreMaxBytes int
	ReadHeaderTimeout  time.Duration
	ReadTimeout        time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
}

func (g *Gateway) RunStreamableHTTP(ctx context.Context, opts HTTPOptions) error {
	normalized, err := normalizeHTTPOptions(opts)
	if err != nil {
		return err
	}
	return g.run(ctx, func(runCtx context.Context) error {
		handler := g.buildStreamableHTTPHandler(normalized)
		mux := http.NewServeMux()
		mux.Handle(normalized.Path, handler)
		if normalized.Path != "/" && !strings.HasSuffix(normalized.Path, "/") {
			mux.Handle(normalized.Path+"/", handler)
		}
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		server := &http.Server{
			Addr:              normalized.Addr,
			Handler:           mux,
			ReadHeaderTimeout: normalized.ReadHeaderTimeout,
			ReadTimeout:       normalized.ReadTimeout,
			IdleTimeout:       normalized.IdleTimeout,
		}

		errCh := make(chan error, 1)
		go func() {
			g.logger.Info("gateway starting (streamable http transport)",
				zap.String("addr", normalized.Addr),
				zap.String("path", normalized.Path),
			)
			var listenErr error
			if normalized.TLSEnabled {
				listenErr = server.ListenAndServeTLS(normalized.TLSCertFile, normalized.TLSKeyFile)
			} else {
				listenErr = server.ListenAndServe()
			}
			if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
				errCh <- listenErr
			}
		}()

		select {
		case <-runCtx.Done():
			shutdownTimeout := normalized.ShutdownTimeout
			if shutdownTimeout <= 0 {
				shutdownTimeout = defaultHTTPShutdownTimeout
			}
			shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancel()
			_ = server.Shutdown(shutdownCtx)
			return runCtx.Err()
		case err := <-errCh:
			return err
		}
	})
}

func normalizeHTTPOptions(opts HTTPOptions) (HTTPOptions, error) {
	if strings.TrimSpace(opts.Addr) == "" {
		return HTTPOptions{}, errors.New("http address is required")
	}
	path := strings.TrimSpace(opts.Path)
	if path == "" {
		path = defaultHTTPPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	opts.Path = path

	if opts.ReadHeaderTimeout <= 0 {
		opts.ReadHeaderTimeout = defaultHTTPReadHeaderTimeout
	}
	if opts.ReadTimeout <= 0 {
		opts.ReadTimeout = defaultHTTPReadTimeout
	}
	if opts.IdleTimeout <= 0 {
		opts.IdleTimeout = defaultHTTPIdleTimeout
	}
	if opts.TLSEnabled {
		if strings.TrimSpace(opts.TLSCertFile) == "" || strings.TrimSpace(opts.TLSKeyFile) == "" {
			return HTTPOptions{}, errors.New("tls enabled but cert or key file is missing")
		}
	}
	return opts, nil
}

func (g *Gateway) buildStreamableHTTPHandler(opts HTTPOptions) http.Handler {
	handler := http.Handler(mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return g.server
	}, &mcp.StreamableHTTPOptions{
		JSONResponse:   opts.JSONResponse,
		SessionTimeout: opts.SessionTimeout,
		EventStore:     buildEventStore(opts),
	}))

	if opts.Token != "" {
		handler = withTokenHeader(handler)
		handler = auth.RequireBearerToken(func(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
			if subtle.ConstantTimeCompare([]byte(token), []byte(opts.Token)) != 1 {
				return nil, auth.ErrInvalidToken
			}
			return &auth.TokenInfo{
				Expiration: time.Now().Add(24 * 365 * 10 * time.Hour),
				UserID:     token,
			}, nil
		}, nil)(handler)
	}

	allowed, allowAll := normalizeAllowedOrigins(opts.AllowedOrigins)
	handler = withOriginCheck(handler, allowed, allowAll)

	return handler
}

func buildEventStore(opts HTTPOptions) mcp.EventStore {
	if !opts.EventStoreEnabled {
		return nil
	}
	store := mcp.NewMemoryEventStore(nil)
	if opts.EventStoreMaxBytes > 0 {
		store.SetMaxBytes(opts.EventStoreMaxBytes)
	}
	return store
}

func normalizeAllowedOrigins(origins []string) (map[string]struct{}, bool) {
	allowed := make(map[string]struct{}, len(origins))
	allowAll := false
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" || trimmed == "*" {
			if trimmed == "*" {
				allowAll = true
			}
			continue
		}
		allowed[trimmed] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, allowAll
	}
	return allowed, allowAll
}

func withTokenHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			if token := strings.TrimSpace(r.Header.Get("X-Mcp-Token")); token != "" {
				r.Header.Set("Authorization", "Bearer "+token)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func withOriginCheck(next http.Handler, allowed map[string]struct{}, allowAll bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !allowAll && allowed == nil {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
		if !allowAll {
			if _, ok := allowed[origin]; !ok {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
			"Authorization",
			"Content-Type",
			"Accept",
			"Mcp-Protocol-Version",
			"Mcp-Session-Id",
			"Last-Event-ID",
			"X-Mcp-Token",
		}, ", "))

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
