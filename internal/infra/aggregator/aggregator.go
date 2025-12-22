package aggregator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"mcpd/internal/domain"
)

type ToolAggregator struct {
	router domain.Router
	specs  map[string]domain.ServerSpec
	cfg    domain.RuntimeConfig
	logger *zap.Logger

	mu         sync.Mutex
	server     *mcp.Server
	registered map[string]struct{}
	ticker     *time.Ticker
	stop       chan struct{}
	started    bool
}

func NewToolAggregator(rt domain.Router, specs map[string]domain.ServerSpec, cfg domain.RuntimeConfig, logger *zap.Logger) *ToolAggregator {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ToolAggregator{
		router:     rt,
		specs:      specs,
		cfg:        cfg,
		logger:     logger.Named("aggregator"),
		registered: make(map[string]struct{}),
		stop:       make(chan struct{}),
	}
}

func (a *ToolAggregator) RegisterServer(s *mcp.Server) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.server = s
}

func (a *ToolAggregator) Start(ctx context.Context) {
	if !a.cfg.ExposeTools {
		return
	}

	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		return
	}
	a.started = true
	a.mu.Unlock()

	if err := a.refresh(ctx); err != nil {
		a.logger.Warn("initial tool refresh failed", zap.Error(err))
	}

	interval := time.Duration(a.cfg.ToolRefreshSeconds) * time.Second
	if interval <= 0 {
		return
	}

	a.mu.Lock()
	if a.ticker != nil {
		a.mu.Unlock()
		return
	}
	a.ticker = time.NewTicker(interval)
	a.mu.Unlock()

	go func() {
		for {
			select {
			case <-a.ticker.C:
				if err := a.refresh(ctx); err != nil {
					a.logger.Warn("tool refresh failed", zap.Error(err))
				}
			case <-a.stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (a *ToolAggregator) Stop() {
	a.mu.Lock()
	if a.ticker != nil {
		a.ticker.Stop()
		a.ticker = nil
	}
	select {
	case <-a.stop:
	default:
		close(a.stop)
	}
	a.mu.Unlock()
}

func (a *ToolAggregator) refresh(ctx context.Context) error {
	a.mu.Lock()
	server := a.server
	a.mu.Unlock()
	if server == nil {
		return errors.New("server is not registered")
	}

	next := make(map[string]struct{})

	for serverType, spec := range a.specs {
		tools, err := a.fetchTools(ctx, serverType)
		if err != nil {
			a.logger.Warn("tool list fetch failed", zap.String("serverType", serverType), zap.Error(err))
			continue
		}

		allowed := allowedTools(spec)
		for _, tool := range tools {
			if tool == nil {
				continue
			}
			if !allowed(tool.Name) {
				continue
			}
			if !isObjectSchema(tool.InputSchema) {
				a.logger.Warn("skip tool with invalid input schema", zap.String("serverType", serverType), zap.String("tool", tool.Name))
				continue
			}

			name := a.namespaceTool(serverType, tool.Name)
			if _, exists := next[name]; exists {
				a.logger.Warn("tool name conflict", zap.String("serverType", serverType), zap.String("tool", tool.Name), zap.String("name", name))
				continue
			}

			toolCopy := *tool
			toolCopy.Name = name

			server.AddTool(&toolCopy, a.forwardToolHandler(serverType, tool.Name))
			next[name] = struct{}{}
		}
	}

	var remove []string
	a.mu.Lock()
	for name := range a.registered {
		if _, ok := next[name]; !ok {
			remove = append(remove, name)
		}
	}
	a.registered = next
	a.mu.Unlock()

	if len(remove) > 0 {
		server.RemoveTools(remove...)
	}

	return nil
}

func (a *ToolAggregator) forwardToolHandler(serverType, toolName string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := &mcp.CallToolParams{
			Name:      toolName,
			Arguments: json.RawMessage(req.Params.Arguments),
		}
		payload, err := buildJSONRPCRequest("tools/call", params)
		if err != nil {
			return errorResult(err), nil
		}

		resp, err := a.router.Route(ctx, serverType, "", payload)
		if err != nil {
			return errorResult(err), nil
		}
		return decodeToolResult(resp)
	}
}

func (a *ToolAggregator) fetchTools(ctx context.Context, serverType string) ([]*mcp.Tool, error) {
	var tools []*mcp.Tool
	cursor := ""

	for {
		params := &mcp.ListToolsParams{Cursor: cursor}
		payload, err := buildJSONRPCRequest("tools/list", params)
		if err != nil {
			return nil, err
		}

		resp, err := a.router.Route(ctx, serverType, "", payload)
		if err != nil {
			return nil, err
		}

		result, err := decodeListToolsResult(resp)
		if err != nil {
			return nil, err
		}
		tools = append(tools, result.Tools...)
		if result.NextCursor == "" {
			break
		}
		cursor = result.NextCursor
	}

	return tools, nil
}

func (a *ToolAggregator) namespaceTool(serverType, toolName string) string {
	if a.cfg.ToolNamespaceStrategy == "flat" {
		return toolName
	}
	return fmt.Sprintf("%s.%s", serverType, toolName)
}

func allowedTools(spec domain.ServerSpec) func(string) bool {
	if len(spec.ExposeTools) == 0 {
		return func(_ string) bool { return true }
	}

	allowed := make(map[string]struct{}, len(spec.ExposeTools))
	for _, name := range spec.ExposeTools {
		allowed[name] = struct{}{}
	}
	return func(name string) bool {
		_, ok := allowed[name]
		return ok
	}
}

func isObjectSchema(schema any) bool {
	if schema == nil {
		return false
	}

	raw, err := json.Marshal(schema)
	if err != nil {
		return false
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return false
	}
	if typ, ok := obj["type"]; ok {
		if val, ok := typ.(string); ok {
			return strings.EqualFold(val, "object")
		}
	}
	return false
}

func buildJSONRPCRequest(method string, params any) (json.RawMessage, error) {
	id, err := jsonrpc.MakeID(fmt.Sprintf("mcpd-%s-%d", method, time.Now().UnixNano()))
	if err != nil {
		return nil, fmt.Errorf("build request id: %w", err)
	}
	rawParams, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}
	req := &jsonrpc.Request{ID: id, Method: method, Params: rawParams}
	wire, err := jsonrpc.EncodeMessage(req)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	return json.RawMessage(wire), nil
}

func decodeListToolsResult(raw json.RawMessage) (*mcp.ListToolsResult, error) {
	resp, err := decodeJSONRPCResponse(raw)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error: %w", resp.Error)
	}

	if len(resp.Result) == 0 {
		return nil, errors.New("tools/list response missing result")
	}

	var result mcp.ListToolsResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("decode tools/list result: %w", err)
	}
	return &result, nil
}

func decodeToolResult(raw json.RawMessage) (*mcp.CallToolResult, error) {
	resp, err := decodeJSONRPCResponse(raw)
	if err != nil {
		return errorResult(err), nil
	}

	if resp.Error != nil {
		return errorResult(resp.Error), nil
	}

	if len(resp.Result) == 0 {
		return errorResult(errors.New("tools/call response missing result")), nil
	}

	var result mcp.CallToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return errorResult(fmt.Errorf("decode tools/call result: %w", err)), nil
	}
	return &result, nil
}

func decodeJSONRPCResponse(raw json.RawMessage) (*jsonrpc.Response, error) {
	msg, err := jsonrpc.DecodeMessage(raw)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	resp, ok := msg.(*jsonrpc.Response)
	if !ok {
		return nil, errors.New("response is not a response message")
	}
	return resp, nil
}

func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("error: %s", err.Error())},
		},
	}
}
