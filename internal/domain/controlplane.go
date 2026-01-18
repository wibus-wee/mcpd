package domain

import (
	"context"
	"encoding/json"
	"time"
)

// ControlPlaneInfo describes control plane identity metadata.
type ControlPlaneInfo struct {
	Name    string
	Version string
	Build   string
}

// ToolDefinition describes a tool exposed by a server.
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

// ToolSnapshot is a versioned snapshot of tools.
type ToolSnapshot struct {
	ETag  string
	Tools []ToolDefinition
}

// ToolTarget identifies the target server for a tool.
type ToolTarget struct {
	ServerType string
	SpecKey    string
	ToolName   string
}

// ResourceDefinition describes a resource exposed by a server.
type ResourceDefinition struct {
	URI         string
	Name        string
	Title       string
	Description string
	MIMEType    string
	Size        int64
	Annotations *Annotations
	Meta        Meta
	SpecKey     string
	ServerName  string
}

// ResourceSnapshot is a versioned snapshot of resources.
type ResourceSnapshot struct {
	ETag      string
	Resources []ResourceDefinition
}

// ResourceTarget identifies the target server for a resource.
type ResourceTarget struct {
	ServerType string
	SpecKey    string
	URI        string
}

// ResourcePage represents a paginated resource snapshot.
type ResourcePage struct {
	Snapshot   ResourceSnapshot
	NextCursor string
}

// PromptDefinition describes a prompt exposed by a server.
type PromptDefinition struct {
	Name        string
	Title       string
	Description string
	Arguments   []PromptArgument
	Meta        Meta
	SpecKey     string
	ServerName  string
}

// PromptSnapshot is a versioned snapshot of prompts.
type PromptSnapshot struct {
	ETag    string
	Prompts []PromptDefinition
}

// PromptTarget identifies the target server for a prompt.
type PromptTarget struct {
	ServerType string
	SpecKey    string
	PromptName string
}

// PromptPage represents a paginated prompt snapshot.
type PromptPage struct {
	Snapshot   PromptSnapshot
	NextCursor string
}

// LogLevel defines the severity for log entries.
type LogLevel string

const (
	// LogLevelDebug represents debug-level logs.
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo represents info-level logs.
	LogLevelInfo LogLevel = "info"
	// LogLevelNotice represents notice-level logs.
	LogLevelNotice LogLevel = "notice"
	// LogLevelWarning represents warning-level logs.
	LogLevelWarning LogLevel = "warning"
	// LogLevelError represents error-level logs.
	LogLevelError LogLevel = "error"
	// LogLevelCritical represents critical-level logs.
	LogLevelCritical LogLevel = "critical"
	// LogLevelAlert represents alert-level logs.
	LogLevelAlert LogLevel = "alert"
	// LogLevelEmergency represents emergency-level logs.
	LogLevelEmergency LogLevel = "emergency"
)

// LogEntry captures a single log entry with structured fields.
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
	Metrics    PoolMetrics
}

// InstanceStatusInfo represents the status of a single server instance
type InstanceStatusInfo struct {
	ID              string
	State           InstanceState
	BusyCount       int
	LastActive      time.Time
	SpawnedAt       time.Time
	HandshakedAt    time.Time
	LastHeartbeatAt time.Time
	LastStartCause  *StartCause
}

// PoolStats contains aggregated statistics for a server pool
type PoolStats struct {
	Total        int
	Ready        int
	Busy         int
	Starting     int
	Initializing int
	Handshaking  int
	Draining     int
	Failed       int
}

// ServerInitStatusSnapshot contains a snapshot of all server initialization statuses
type ServerInitStatusSnapshot struct {
	Statuses    []ServerInitStatus
	GeneratedAt time.Time
}

// ControlPlaneInfoProvider exposes basic control plane metadata.
type ControlPlaneInfoProvider interface {
	Info(ctx context.Context) (ControlPlaneInfo, error)
}

// ControlPlaneRegistry manages caller registration and monitoring.
type ControlPlaneRegistry interface {
	RegisterCaller(ctx context.Context, caller string, pid int) (string, error)
	UnregisterCaller(ctx context.Context, caller string) error
	ListActiveCallers(ctx context.Context) ([]ActiveCaller, error)
	WatchActiveCallers(ctx context.Context) (<-chan ActiveCallerSnapshot, error)
}

// ControlPlaneDiscovery exposes tools, resources, and prompts.
type ControlPlaneDiscovery interface {
	ListTools(ctx context.Context, caller string) (ToolSnapshot, error)
	ListToolsAllProfiles(ctx context.Context) (ToolSnapshot, error)
	ListToolCatalog(ctx context.Context) (ToolCatalogSnapshot, error)
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
}

// ServerInitStatusReader provides server initialization status snapshots.
type ServerInitStatusReader interface {
	GetServerInitStatus(ctx context.Context) ([]ServerInitStatus, error)
}

// ControlPlaneObservability exposes runtime status and log streaming.
type ControlPlaneObservability interface {
	StreamLogs(ctx context.Context, caller string, minLevel LogLevel) (<-chan LogEntry, error)
	StreamLogsAllProfiles(ctx context.Context, minLevel LogLevel) (<-chan LogEntry, error)
	GetPoolStatus(ctx context.Context) ([]PoolInfo, error)
	ServerInitStatusReader
	RetryServerInit(ctx context.Context, specKey string) error
	WatchRuntimeStatus(ctx context.Context, caller string) (<-chan RuntimeStatusSnapshot, error)
	WatchRuntimeStatusAllProfiles(ctx context.Context) (<-chan RuntimeStatusSnapshot, error)
	WatchServerInitStatus(ctx context.Context, caller string) (<-chan ServerInitStatusSnapshot, error)
	WatchServerInitStatusAllProfiles(ctx context.Context) (<-chan ServerInitStatusSnapshot, error)
}

// ControlPlaneBootstrap exposes bootstrap status.
type ControlPlaneBootstrap interface {
	GetBootstrapProgress(ctx context.Context) (BootstrapProgress, error)
	WatchBootstrapProgress(ctx context.Context) (<-chan BootstrapProgress, error)
}

// ControlPlaneAutomation exposes automatic tool filtering and execution.
type ControlPlaneAutomation interface {
	AutomaticMCP(ctx context.Context, caller string, params AutomaticMCPParams) (AutomaticMCPResult, error)
	AutomaticEval(ctx context.Context, caller string, params AutomaticEvalParams) (json.RawMessage, error)
	IsSubAgentEnabled() bool
	IsSubAgentEnabledForCaller(caller string) bool
}

// ControlPlaneStore exposes profile storage access.
type ControlPlaneStore interface {
	GetProfileStore() ProfileStore
}

// ControlPlane groups all control plane capabilities.
type ControlPlane interface {
	ControlPlaneInfoProvider
	ControlPlaneRegistry
	ControlPlaneDiscovery
	ControlPlaneObservability
	ControlPlaneBootstrap
	ControlPlaneAutomation
	ControlPlaneStore
}
