package aggregator

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"mcpd/internal/domain"
)

func TestToolAggregator_RegistersPrefixedTool(t *testing.T) {
	ctx := context.Background()
	router := &fakeRouter{
		tools: []*mcp.Tool{
			{
				Name:        "echo",
				Description: "echo input",
				InputSchema: map[string]any{"type": "object"},
			},
		},
		callResult: &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
		},
	}

	specs := map[string]domain.ServerSpec{
		"echo": {Name: "echo"},
	}
	cfg := domain.RuntimeConfig{
		ExposeTools:           true,
		ToolNamespaceStrategy: "prefix",
		ToolRefreshSeconds:    0,
	}

	agg := NewToolAggregator(router, specs, cfg, zap.NewNop())
	server := mcp.NewServer(&mcp.Implementation{Name: "mcpd", Version: "0.1.0"}, nil)
	agg.RegisterServer(server)
	agg.Start(ctx)
	defer agg.Stop()

	_, session := connectClient(t, ctx, server)
	defer session.Close()

	res, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, res.Tools, 1)
	require.Equal(t, "echo.echo", res.Tools[0].Name)

	callRes, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "echo.echo", Arguments: map[string]any{}})
	require.NoError(t, err)
	require.Len(t, callRes.Content, 1)
	require.Equal(t, "ok", callRes.Content[0].(*mcp.TextContent).Text)

	require.Equal(t, "tools/call", router.lastMethod)
	require.Equal(t, "echo", router.lastServerType)
}

func TestToolAggregator_RespectsExposeToolsAllowlist(t *testing.T) {
	ctx := context.Background()
	router := &fakeRouter{
		tools: []*mcp.Tool{
			{Name: "echo", InputSchema: map[string]any{"type": "object"}},
			{Name: "skip", InputSchema: map[string]any{"type": "object"}},
		},
		callResult: &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}},
	}

	specs := map[string]domain.ServerSpec{
		"echo": {
			Name:        "echo",
			ExposeTools: []string{"echo"},
		},
	}
	cfg := domain.RuntimeConfig{ExposeTools: true, ToolNamespaceStrategy: "prefix"}

	agg := NewToolAggregator(router, specs, cfg, zap.NewNop())
	server := mcp.NewServer(&mcp.Implementation{Name: "mcpd", Version: "0.1.0"}, nil)
	agg.RegisterServer(server)
	agg.Start(ctx)
	defer agg.Stop()

	_, session := connectClient(t, ctx, server)
	defer session.Close()

	res, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	require.Len(t, res.Tools, 1)
	require.Equal(t, "echo.echo", res.Tools[0].Name)
}

type fakeRouter struct {
	tools          []*mcp.Tool
	callResult     *mcp.CallToolResult
	lastMethod     string
	lastServerType string
}

func (f *fakeRouter) Route(ctx context.Context, serverType, routingKey string, payload json.RawMessage) (json.RawMessage, error) {
	msg, err := jsonrpc.DecodeMessage(payload)
	if err != nil {
		return nil, err
	}
	req, ok := msg.(*jsonrpc.Request)
	if !ok {
		return nil, errors.New("invalid jsonrpc request")
	}
	f.lastMethod = req.Method
	f.lastServerType = serverType

	switch req.Method {
	case "tools/list":
		return encodeResponse(req.ID, &mcp.ListToolsResult{Tools: f.tools})
	case "tools/call":
		return encodeResponse(req.ID, f.callResult)
	default:
		return nil, nil
	}
}

func encodeResponse(id jsonrpc.ID, result any) (json.RawMessage, error) {
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	resp := &jsonrpc.Response{ID: id, Result: raw}
	wire, err := jsonrpc.EncodeMessage(resp)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(wire), nil
}

func connectClient(t *testing.T, ctx context.Context, server *mcp.Server) (*mcp.Client, *mcp.ClientSession) {
	t.Helper()
	ct, st := mcp.NewInMemoryTransports()
	_, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	return client, session
}
