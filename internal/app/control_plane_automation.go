package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcpd/internal/domain"
	"mcpd/internal/infra/mcpcodec"
)

const (
	defaultAutomaticMCPSessionTTL = 30 * time.Minute
	defaultAutomaticMCPCacheSize  = 10000
)

type automationService struct {
	state     *controlPlaneState
	registry  *callerRegistry
	discovery *discoveryService
	subAgent  domain.SubAgent
	cache     *domain.SessionCache
}

func newAutomationService(state *controlPlaneState, registry *callerRegistry, discovery *discoveryService) *automationService {
	return &automationService{
		state:     state,
		registry:  registry,
		discovery: discovery,
		cache:     domain.NewSessionCache(defaultAutomaticMCPSessionTTL, defaultAutomaticMCPCacheSize),
	}
}

func (a *automationService) SetSubAgent(agent domain.SubAgent) {
	a.subAgent = agent
}

func (a *automationService) IsSubAgentEnabled() bool {
	return a.subAgent != nil
}

func (a *automationService) IsSubAgentEnabledForCaller(caller string) bool {
	if a.subAgent == nil {
		return false
	}

	profile, err := a.registry.resolveProfile(caller)
	if err != nil {
		return false
	}

	if a.state.profileStore.Profiles == nil {
		return false
	}
	profileData, ok := a.state.profileStore.Profiles[profile.name]
	if !ok {
		return false
	}

	return profileData.Catalog.SubAgent.Enabled
}

func (a *automationService) AutomaticMCP(ctx context.Context, caller string, params domain.AutomaticMCPParams) (domain.AutomaticMCPResult, error) {
	profile, err := a.registry.resolveProfile(caller)
	if err != nil {
		return domain.AutomaticMCPResult{}, err
	}

	if a.subAgent != nil {
		return a.subAgent.SelectToolsForCaller(ctx, caller, params)
	}

	return a.fallbackAutomaticMCP(caller, profile, params)
}

func (a *automationService) fallbackAutomaticMCP(caller string, profile *profileRuntime, params domain.AutomaticMCPParams) (domain.AutomaticMCPResult, error) {
	if profile.tools == nil {
		return domain.AutomaticMCPResult{}, nil
	}

	snapshot := profile.tools.Snapshot()
	sessionKey := domain.AutomaticMCPSessionKey(caller, params.SessionID)

	toolsToSend := make([]domain.ToolDefinition, 0, len(snapshot.Tools))
	sentSchemas := make(map[string]string)
	for _, tool := range snapshot.Tools {
		hash := mcpcodec.HashToolDefinition(tool)
		shouldSend := params.ForceRefresh || a.cache.NeedsFull(sessionKey, tool.Name, hash)
		if !shouldSend {
			continue
		}

		toolsToSend = append(toolsToSend, domain.CloneToolDefinition(tool))
		sentSchemas[tool.Name] = hash
	}

	a.cache.Update(sessionKey, sentSchemas)

	return domain.AutomaticMCPResult{
		ETag:           snapshot.ETag,
		Tools:          toolsToSend,
		TotalAvailable: len(snapshot.Tools),
		Filtered:       len(toolsToSend),
	}, nil
}

func (a *automationService) AutomaticEval(ctx context.Context, caller string, params domain.AutomaticEvalParams) (json.RawMessage, error) {
	tool, err := a.getToolDefinition(caller, params.ToolName)
	if err != nil {
		return nil, err
	}

	if err := validateToolArguments(tool, params.Arguments); err != nil {
		result, buildErr := buildAutomaticEvalSchemaError(tool, err)
		if buildErr != nil {
			return nil, buildErr
		}
		return result, nil
	}

	return a.discovery.CallTool(ctx, caller, params.ToolName, params.Arguments, params.RoutingKey)
}

func (a *automationService) getToolDefinition(caller, name string) (domain.ToolDefinition, error) {
	profile, err := a.registry.resolveProfile(caller)
	if err != nil {
		return domain.ToolDefinition{}, err
	}
	if profile.tools == nil {
		return domain.ToolDefinition{}, domain.ErrToolNotFound
	}

	snapshot := profile.tools.Snapshot()
	for _, tool := range snapshot.Tools {
		if tool.Name == name {
			return tool, nil
		}
	}
	return domain.ToolDefinition{}, domain.ErrToolNotFound
}

func validateToolArguments(tool domain.ToolDefinition, args json.RawMessage) error {
	if tool.InputSchema == nil {
		return nil
	}

	var schema jsonschema.Schema
	schemaJSON, err := json.Marshal(tool.InputSchema)
	if err != nil {
		return fmt.Errorf("encode tool input schema: %w", err)
	}
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return fmt.Errorf("decode tool input schema: %w", err)
	}

	resolved, err := schema.Resolve(nil)
	if err != nil {
		return fmt.Errorf("resolve tool input schema: %w", err)
	}

	var payload any
	if len(args) == 0 {
		payload = map[string]any{}
	} else if err := json.Unmarshal(args, &payload); err != nil {
		return fmt.Errorf("decode tool arguments: %w", err)
	}

	if err := resolved.Validate(payload); err != nil {
		return fmt.Errorf("invalid tool arguments: %w", err)
	}
	return nil
}

func buildAutomaticEvalSchemaError(tool domain.ToolDefinition, err error) (json.RawMessage, error) {
	toolJSON, marshalErr := mcpcodec.MarshalToolDefinition(tool)
	if marshalErr != nil {
		return nil, marshalErr
	}
	payload := struct {
		Error      string          `json:"error"`
		ToolSchema json.RawMessage `json:"toolSchema"`
	}{
		Error:      err.Error(),
		ToolSchema: toolJSON,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	result := mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(payloadJSON)},
		},
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
