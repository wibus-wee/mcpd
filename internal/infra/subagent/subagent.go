package subagent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"mcpd/internal/domain"
)

const (
	defaultSessionTTL     = 30 * time.Minute
	defaultMaxCacheSize   = 10000
	defaultMaxToolsReturn = 50
)

// EinoSubAgent filters and proxies tool calls using an LLM.
type EinoSubAgent struct {
	config       domain.SubAgentConfig
	model        model.ChatModel
	cache        *domain.SessionCache
	controlPlane controlPlaneProvider
	logger       *zap.Logger
}

// controlPlaneProvider provides access to profile-specific tool snapshots.
type controlPlaneProvider interface {
	GetToolSnapshotForCaller(caller string) (domain.ToolSnapshot, error)
}

// NewEinoSubAgent creates a new SubAgent instance.
func NewEinoSubAgent(
	ctx context.Context,
	config domain.SubAgentConfig,
	controlPlane controlPlaneProvider,
	logger *zap.Logger,
) (*EinoSubAgent, error) {
	// Initialize LLM model based on config
	chatModel, err := initializeModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("initialize model: %w", err)
	}

	return &EinoSubAgent{
		config:       config,
		model:        chatModel,
		cache:        domain.NewSessionCache(defaultSessionTTL, defaultMaxCacheSize),
		controlPlane: controlPlane,
		logger:       logger.Named("subagent"),
	}, nil
}

// SelectToolsForCaller filters tools based on the query using LLM reasoning
// and applies deduplication based on session cache.
func (s *EinoSubAgent) SelectToolsForCaller(
	ctx context.Context,
	callerID string,
	params domain.AutomaticMCPParams,
) (domain.AutomaticMCPResult, error) {
	snapshot, err := s.controlPlane.GetToolSnapshotForCaller(callerID)
	if err != nil {
		return domain.AutomaticMCPResult{}, fmt.Errorf("get tool snapshot: %w", err)
	}

	if len(snapshot.Tools) == 0 {
		return domain.AutomaticMCPResult{ETag: snapshot.ETag}, nil
	}

	// Build tool summaries and hash map
	summaries, hashMap := s.buildToolSummaries(snapshot.Tools)

	// Use LLM to select relevant tools (if query provided)
	var selectedNames []string
	if params.Query != "" {
		selectedNames, err = s.filterWithLLM(ctx, params.Query, summaries)
		if err != nil {
			// Fallback: return all tools if LLM fails
			s.logger.Warn("LLM filtering failed, returning all tools", zap.Error(err))
			selectedNames = allToolNames(summaries)
		}
	} else {
		// No query - return all tools
		selectedNames = allToolNames(summaries)
	}

	// Apply max tools limit
	maxTools := s.config.MaxToolsPerRequest
	if maxTools <= 0 {
		maxTools = defaultMaxToolsReturn
	}
	if len(selectedNames) > maxTools {
		selectedNames = selectedNames[:maxTools]
	}

	sessionKey := domain.AutomaticMCPSessionKey(callerID, params.SessionID)

	shouldSend := make(map[string]bool, len(selectedNames))
	for _, name := range selectedNames {
		shouldSend[name] = params.ForceRefresh || s.cache.NeedsFull(sessionKey, name, hashMap[name])
	}

	toolsToSend, sentSchemas := s.buildToolPayloads(snapshot.Tools, selectedNames, hashMap, shouldSend)
	s.cache.Update(sessionKey, sentSchemas)

	return domain.AutomaticMCPResult{
		ETag:           snapshot.ETag,
		Tools:          toolsToSend,
		TotalAvailable: len(snapshot.Tools),
		Filtered:       len(toolsToSend),
	}, nil
}

// InvalidateSession clears the session cache for a caller.
func (s *EinoSubAgent) InvalidateSession(callerID string) {
	s.cache.Invalidate(callerID)
}

// Close shuts down the SubAgent and releases resources.
func (s *EinoSubAgent) Close() error {
	// Clean up any resources
	return nil
}

// toolSummary contains minimal info for LLM filtering.
type toolSummary struct {
	Name        string
	Description string
	ParamCount  int
}

// buildToolSummaries creates summaries for LLM and computes schema hashes.
func (s *EinoSubAgent) buildToolSummaries(tools []domain.ToolDefinition) ([]toolSummary, map[string]string) {
	summaries := make([]toolSummary, 0, len(tools))
	hashMap := make(map[string]string, len(tools))

	for _, t := range tools {
		var toolObj struct {
			Description string `json:"description"`
			InputSchema struct {
				Properties map[string]interface{} `json:"properties"`
			} `json:"inputSchema"`
		}
		_ = json.Unmarshal(t.ToolJSON, &toolObj)

		summaries = append(summaries, toolSummary{
			Name:        t.Name,
			Description: toolObj.Description,
			ParamCount:  len(toolObj.InputSchema.Properties),
		})

		// Compute hash of full schema
		hash := sha256.Sum256(t.ToolJSON)
		hashMap[t.Name] = hex.EncodeToString(hash[:])
	}

	return summaries, hashMap
}

// filterWithLLM uses the LLM to select relevant tools for the query.
func (s *EinoSubAgent) filterWithLLM(
	ctx context.Context,
	query string,
	summaries []toolSummary,
) ([]string, error) {
	if len(summaries) == 0 {
		return nil, nil
	}

	// Build prompt for tool selection
	prompt := s.buildFilterPrompt(query, summaries)

	messages := []*schema.Message{
		schema.SystemMessage(defaultFilterSystemPrompt),
		schema.UserMessage(prompt),
	}

	response, err := s.model.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generate: %w", err)
	}

	// Parse response to extract tool names
	return s.parseSelectedTools(response.Content, summaries)
}

// buildFilterPrompt creates the prompt for tool selection.
func (s *EinoSubAgent) buildFilterPrompt(query string, summaries []toolSummary) string {
	var sb strings.Builder
	sb.WriteString("User task: ")
	sb.WriteString(query)
	sb.WriteString("\n\nAvailable tools:\n")

	for _, t := range summaries {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", t.Name, t.Description))
	}

	sb.WriteString("\nSelect only the tools that are directly relevant to completing this task.")
	return sb.String()
}

// parseSelectedTools extracts tool names from LLM response.
func (s *EinoSubAgent) parseSelectedTools(response string, summaries []toolSummary) ([]string, error) {
	// Build a map of valid tool names for validation
	validNames := make(map[string]bool, len(summaries))
	for _, t := range summaries {
		validNames[t.Name] = true
	}

	// Try to parse as JSON array first
	var jsonNames []string
	if err := json.Unmarshal([]byte(response), &jsonNames); err == nil {
		result := make([]string, 0, len(jsonNames))
		for _, name := range jsonNames {
			if validNames[name] {
				result = append(result, name)
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}

	// Fallback: scan for tool names in response text
	result := make([]string, 0)
	for name := range validNames {
		if strings.Contains(response, name) {
			result = append(result, name)
		}
	}

	if len(result) == 0 {
		// If nothing found, return all tools
		return allToolNames(summaries), nil
	}

	return result, nil
}

// buildToolPayloads builds the tool payload list with deduplication applied.
func (s *EinoSubAgent) buildToolPayloads(
	tools []domain.ToolDefinition,
	selectedNames []string,
	hashMap map[string]string,
	shouldSend map[string]bool,
) ([]json.RawMessage, map[string]string) {
	// Build lookup for tools
	toolMap := make(map[string]domain.ToolDefinition, len(tools))
	for _, t := range tools {
		toolMap[t.Name] = t
	}

	result := make([]json.RawMessage, 0, len(selectedNames))
	sentSchemas := make(map[string]string)
	for _, name := range selectedNames {
		t, ok := toolMap[name]
		if !ok {
			continue
		}
		if !shouldSend[name] {
			continue
		}

		raw := make([]byte, len(t.ToolJSON))
		copy(raw, t.ToolJSON)
		result = append(result, raw)
		sentSchemas[name] = hashMap[name]
	}

	return result, sentSchemas
}

// allToolNames extracts all tool names from summaries.
func allToolNames(summaries []toolSummary) []string {
	names := make([]string, len(summaries))
	for i, t := range summaries {
		names[i] = t.Name
	}
	return names
}

const defaultFilterSystemPrompt = `You are a tool selection assistant. Given a user task and a list of available tools, select only the tools that are relevant to completing the task.

Output format: JSON array of tool names that are relevant.
Example: ["tool1", "tool2"]

Be selective - only include tools that are directly useful for the given task. Do not include tools that are only tangentially related.`
