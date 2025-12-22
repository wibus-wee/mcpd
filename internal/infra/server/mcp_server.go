package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"mcpd/internal/domain"
	"mcpd/internal/infra/aggregator"
	"mcpd/internal/infra/telemetry"
)

type routeArgs struct {
	ServerType string          `json:"serverType"`
	RoutingKey string          `json:"routingKey,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

// Run starts an MCP server over stdio using go-sdk and routes tool calls to the provided router.
func Run(ctx context.Context, rt domain.Router, cfg domain.RuntimeConfig, agg *aggregator.ToolAggregator, logSink *telemetry.MCPLogSink, logger *zap.Logger) error {
	s, stop := newServer(ctx, rt, cfg, agg, logSink, logger)
	defer stop()

	logger.Info("mcp server starting (stdio transport)")
	return s.Run(ctx, &mcp.StdioTransport{})
}

func newServer(ctx context.Context, rt domain.Router, cfg domain.RuntimeConfig, agg *aggregator.ToolAggregator, logSink *telemetry.MCPLogSink, logger *zap.Logger) (*mcp.Server, func()) {
	if logger == nil {
		logger = zap.NewNop()
	}

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "mcpd",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "route",
		Description: "Route a JSON-RPC payload to a configured MCP server type",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args routeArgs) (*mcp.CallToolResult, any, error) {
		if args.ServerType == "" {
			return errorResult("serverType is required"), nil, nil
		}
		if len(args.Payload) == 0 {
			return errorResult("payload is required"), nil, nil
		}
		resp, err := rt.Route(ctx, args.ServerType, args.RoutingKey, args.Payload)
		if err != nil {
			logger.Warn("route tool failed", zap.String("serverType", args.ServerType), zap.Error(err))
			return errorResult(err.Error()), nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resp)},
			},
		}, nil, nil
	})

	stop := func() {}
	if agg != nil && cfg.ExposeTools {
		agg.RegisterServer(s)
		agg.Start(ctx)
		stop = agg.Stop
	}
	if logSink != nil {
		logSink.SetServer(s)
		previousStop := stop
		stop = func() {
			previousStop()
			logSink.SetServer(nil)
		}
	}

	return s, stop
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("error: %s", msg)},
		},
	}
}
