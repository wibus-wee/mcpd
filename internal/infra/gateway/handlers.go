package gateway

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	controlv1 "mcpv/pkg/api/control/v1"
)

func (g *Gateway) toolHandler(name string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args json.RawMessage
		if req != nil && req.Params != nil {
			args = json.RawMessage(req.Params.Arguments)
		}
		resp, err := g.callTool(ctx, name, args)
		if err != nil {
			return nil, err
		}
		var result mcp.CallToolResult
		if err := json.Unmarshal(resp.GetResultJson(), &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
}

func (g *Gateway) promptHandler(name string) mcp.PromptHandler {
	return func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		var args json.RawMessage
		if req != nil && req.Params != nil {
			raw, err := json.Marshal(req.Params.Arguments)
			if err != nil {
				return nil, err
			}
			args = raw
		}
		resp, err := g.getPrompt(ctx, name, args)
		if err != nil {
			return nil, err
		}
		var result mcp.GetPromptResult
		if err := json.Unmarshal(resp.GetResultJson(), &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
}

func (g *Gateway) resourceHandler(uri string) mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		targetURI := uri
		if req != nil && req.Params != nil && req.Params.URI != "" {
			targetURI = req.Params.URI
		}
		resp, err := g.readResource(ctx, targetURI)
		if err != nil {
			return nil, err
		}
		var result mcp.ReadResourceResult
		if err := json.Unmarshal(resp.GetResultJson(), &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
}

// automaticMCPHandler handles the mcpv.automatic_mcp tool call.
func (g *Gateway) automaticMCPHandler() mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params struct {
			Query        string `json:"query"`
			SessionID    string `json:"sessionId"`
			ForceRefresh bool   `json:"forceRefresh"`
		}
		if req != nil && req.Params != nil && req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return nil, err
			}
		}

		client, err := g.clients.get(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := client.Control().AutomaticMCP(ctx, &controlv1.AutomaticMCPRequest{
			Caller:       g.caller,
			Query:        params.Query,
			SessionId:    params.SessionID,
			ForceRefresh: params.ForceRefresh,
		})
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				if regErr := g.registerCaller(ctx); regErr == nil {
					resp, err = client.Control().AutomaticMCP(ctx, &controlv1.AutomaticMCPRequest{
						Caller:       g.caller,
						Query:        params.Query,
						SessionId:    params.SessionID,
						ForceRefresh: params.ForceRefresh,
					})
				}
			}
			if err != nil {
				return nil, err
			}
		}

		type automaticMCPResult struct {
			ETag           string            `json:"etag"`
			Tools          []json.RawMessage `json:"tools"`
			TotalAvailable int               `json:"totalAvailable"`
			Filtered       int               `json:"filtered"`
		}

		tools := make([]json.RawMessage, 0, len(resp.GetToolsJson()))
		for _, raw := range resp.GetToolsJson() {
			if len(raw) == 0 {
				continue
			}
			tools = append(tools, json.RawMessage(raw))
		}

		resultJSON, err := json.Marshal(automaticMCPResult{
			ETag:           resp.GetEtag(),
			Tools:          tools,
			TotalAvailable: int(resp.GetTotalAvailable()),
			Filtered:       int(resp.GetFiltered()),
		})
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resultJSON)},
			},
		}, nil
	}
}

// automaticEvalHandler handles the mcpv.automatic_eval tool call.
func (g *Gateway) automaticEvalHandler() mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params struct {
			ToolName   string          `json:"toolName"`
			Arguments  json.RawMessage `json:"arguments"`
			RoutingKey string          `json:"routingKey"`
		}
		if req != nil && req.Params != nil && req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
				return nil, err
			}
		}

		if params.ToolName == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "toolName is required"},
				},
			}, nil
		}

		client, err := g.clients.get(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := client.Control().AutomaticEval(ctx, &controlv1.AutomaticEvalRequest{
			Caller:        g.caller,
			ToolName:      params.ToolName,
			ArgumentsJson: params.Arguments,
			RoutingKey:    params.RoutingKey,
		})
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				if regErr := g.registerCaller(ctx); regErr == nil {
					resp, err = client.Control().AutomaticEval(ctx, &controlv1.AutomaticEvalRequest{
						Caller:        g.caller,
						ToolName:      params.ToolName,
						ArgumentsJson: params.Arguments,
						RoutingKey:    params.RoutingKey,
					})
				}
			}
			if err != nil {
				return nil, err
			}
		}

		if resp == nil || len(resp.GetResultJson()) == 0 {
			return nil, errors.New("empty automatic_eval response")
		}

		var result mcp.CallToolResult
		if err := json.Unmarshal(resp.GetResultJson(), &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
}
