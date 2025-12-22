package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"mcpd/internal/infra/router"
)

type routeArgs struct {
	ServerType string          `json:"serverType"`
	RoutingKey string          `json:"routingKey,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

// Run starts an MCP server over stdio using go-sdk and routes tool calls to the provided router.
func Run(ctx context.Context, rt *router.BasicRouter, logger *zap.Logger) error {
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
			return errorResult(err.Error()), nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resp)},
			},
		}, nil, nil
	})

	logger.Info("mcp server starting (stdio transport)")
	return s.Run(ctx, &mcp.StdioTransport{})
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("error: %s", msg)},
		},
	}
}
