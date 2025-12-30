package transport

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"mcpd/internal/domain"
)

type MCPTransport struct {
	logger            *zap.Logger
	listChangeEmitter domain.ListChangeEmitter
}

type MCPTransportOptions struct {
	Logger            *zap.Logger
	ListChangeEmitter domain.ListChangeEmitter
}

func NewMCPTransport(opts MCPTransportOptions) *MCPTransport {
	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MCPTransport{
		logger:            logger,
		listChangeEmitter: opts.ListChangeEmitter,
	}
}

func (t *MCPTransport) Connect(ctx context.Context, specKey string, spec domain.ServerSpec, streams domain.IOStreams) (domain.Conn, error) {
	if streams.Reader == nil || streams.Writer == nil {
		return nil, errors.New("streams are required")
	}

	transport := &mcp.IOTransport{
		Reader: streams.Reader,
		Writer: streams.Writer,
	}
	mcpConn, err := transport.Connect(ctx)
	if err != nil {
		_ = streams.Reader.Close()
		_ = streams.Writer.Close()
		return nil, fmt.Errorf("connect io transport: %w", err)
	}

	return newClientConn(mcpConn, clientConnOptions{
		Logger:            t.logger.Named("mcp_conn"),
		ListChangeEmitter: t.listChangeEmitter,
		ServerType:        spec.Name,
		SpecKey:           specKey,
	}), nil
}
