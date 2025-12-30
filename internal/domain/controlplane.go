package domain

import (
	"context"
	"encoding/json"
	"time"
)

type ControlPlaneInfo struct {
	Name    string
	Version string
	Build   string
}

type ToolDefinition struct {
	Name         string
	Description  string
	InputSchema  any
	OutputSchema any
	Title        string
	Annotations  *ToolAnnotations
	Meta         Meta
	SpecKey      string
	ServerName   string
}

type ToolSnapshot struct {
	ETag  string
	Tools []ToolDefinition
}

type ToolTarget struct {
	ServerType string
	SpecKey    string
	ToolName   string
}

type ResourceDefinition struct {
	URI         string
	Name        string
	Title       string
	Description string
	MIMEType    string
	Size        int64
	Annotations *Annotations
	Meta        Meta
}

type ResourceSnapshot struct {
	ETag      string
	Resources []ResourceDefinition
}

type ResourceTarget struct {
	ServerType string
	SpecKey    string
	URI        string
}

type ResourcePage struct {
	Snapshot   ResourceSnapshot
	NextCursor string
}

type PromptDefinition struct {
	Name        string
	Title       string
	Description string
	Arguments   []PromptArgument
	Meta        Meta
}

type PromptSnapshot struct {
	ETag    string
	Prompts []PromptDefinition
}

type PromptTarget struct {
	ServerType string
	SpecKey    string
	PromptName string
}

type PromptPage struct {
	Snapshot   PromptSnapshot
	NextCursor string
}

type LogLevel string

const (
	LogLevelDebug     LogLevel = "debug"
	LogLevelInfo      LogLevel = "info"
	LogLevelNotice    LogLevel = "notice"
	LogLevelWarning   LogLevel = "warning"
	LogLevelError     LogLevel = "error"
	LogLevelCritical  LogLevel = "critical"
	LogLevelAlert     LogLevel = "alert"
	LogLevelEmergency LogLevel = "emergency"
)

type LogEntry struct {
	Logger    string
	Level     LogLevel
	Timestamp time.Time
	Data      map[string]any
}

// ActiveCaller represents a registered caller in the control plane.
type ActiveCaller struct {
	Caller        string
	PID           int
	Profile       string
	LastHeartbeat time.Time
}

// ActiveCallerSnapshot contains a snapshot of active callers.
type ActiveCallerSnapshot struct {
	Callers     []ActiveCaller
	GeneratedAt time.Time
}

// RuntimeStatusSnapshot contains a snapshot of all server runtime statuses
type RuntimeStatusSnapshot struct {
	ETag        string
	Statuses    []ServerRuntimeStatus
	GeneratedAt time.Time
}

// ServerRuntimeStatus contains the runtime status of a server and its instances
type ServerRuntimeStatus struct {
	SpecKey    string
	ServerName string
	Instances  []InstanceStatusInfo
	Stats      PoolStats
}

// InstanceStatusInfo represents the status of a single server instance
type InstanceStatusInfo struct {
	ID         string
	State      InstanceState
	BusyCount  int
	LastActive time.Time
}

// PoolStats contains aggregated statistics for a server pool
type PoolStats struct {
	Total    int
	Ready    int
	Busy     int
	Starting int
	Draining int
	Failed   int
}

// ServerInitStatusSnapshot contains a snapshot of all server initialization statuses
type ServerInitStatusSnapshot struct {
	Statuses    []ServerInitStatus
	GeneratedAt time.Time
}

type ControlPlane interface {
	Info(ctx context.Context) (ControlPlaneInfo, error)
	RegisterCaller(ctx context.Context, caller string, pid int) (string, error)
	UnregisterCaller(ctx context.Context, caller string) error
	ListActiveCallers(ctx context.Context) ([]ActiveCaller, error)
	WatchActiveCallers(ctx context.Context) (<-chan ActiveCallerSnapshot, error)
	ListTools(ctx context.Context, caller string) (ToolSnapshot, error)
	ListToolsAllProfiles(ctx context.Context) (ToolSnapshot, error)
	WatchTools(ctx context.Context, caller string) (<-chan ToolSnapshot, error)
	CallTool(ctx context.Context, caller, name string, args json.RawMessage, routingKey string) (json.RawMessage, error)
	CallToolAllProfiles(ctx context.Context, name string, args json.RawMessage, routingKey, specKey string) (json.RawMessage, error)
	ListResources(ctx context.Context, caller string, cursor string) (ResourcePage, error)
	ListResourcesAllProfiles(ctx context.Context, cursor string) (ResourcePage, error)
	WatchResources(ctx context.Context, caller string) (<-chan ResourceSnapshot, error)
	ReadResource(ctx context.Context, caller, uri string) (json.RawMessage, error)
	ReadResourceAllProfiles(ctx context.Context, uri, specKey string) (json.RawMessage, error)
	ListPrompts(ctx context.Context, caller string, cursor string) (PromptPage, error)
	ListPromptsAllProfiles(ctx context.Context, cursor string) (PromptPage, error)
	WatchPrompts(ctx context.Context, caller string) (<-chan PromptSnapshot, error)
	GetPrompt(ctx context.Context, caller, name string, args json.RawMessage) (json.RawMessage, error)
	GetPromptAllProfiles(ctx context.Context, name string, args json.RawMessage, specKey string) (json.RawMessage, error)
	StreamLogs(ctx context.Context, caller string, minLevel LogLevel) (<-chan LogEntry, error)
	StreamLogsAllProfiles(ctx context.Context, minLevel LogLevel) (<-chan LogEntry, error)
	GetProfileStore() ProfileStore
	GetPoolStatus(ctx context.Context) ([]PoolInfo, error)
	GetServerInitStatus(ctx context.Context) ([]ServerInitStatus, error)
	RetryServerInit(ctx context.Context, specKey string) error
	WatchRuntimeStatus(ctx context.Context, caller string) (<-chan RuntimeStatusSnapshot, error)
	WatchRuntimeStatusAllProfiles(ctx context.Context) (<-chan RuntimeStatusSnapshot, error)
	WatchServerInitStatus(ctx context.Context, caller string) (<-chan ServerInitStatusSnapshot, error)
	WatchServerInitStatusAllProfiles(ctx context.Context) (<-chan ServerInitStatusSnapshot, error)

	// AutomaticMCP returns filtered tool metadata based on caller profile and query.
	// When SubAgent is enabled, uses LLM to filter tools by relevance.
	// Uses session-based hash tracking to minimize schema resending.
	AutomaticMCP(ctx context.Context, caller string, params AutomaticMCPParams) (AutomaticMCPResult, error)

	// AutomaticEval proxies a tool call to the actual MCP tool implementation.
	AutomaticEval(ctx context.Context, caller string, params AutomaticEvalParams) (json.RawMessage, error)

	// IsSubAgentEnabled returns whether the SubAgent infrastructure is available.
	IsSubAgentEnabled() bool

	// IsSubAgentEnabledForCaller returns whether SubAgent is enabled for the caller's profile.
	IsSubAgentEnabledForCaller(caller string) bool
}
