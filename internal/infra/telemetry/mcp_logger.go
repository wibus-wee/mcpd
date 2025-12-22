package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap/zapcore"
)

type MCPLogSink struct {
	mu         sync.RWMutex
	server     *mcp.Server
	loggerName string
	minLevel   zapcore.Level
}

func NewMCPLogSink(loggerName string, minLevel zapcore.Level) *MCPLogSink {
	return &MCPLogSink{
		loggerName: loggerName,
		minLevel:   minLevel,
	}
}

func (s *MCPLogSink) SetServer(server *mcp.Server) {
	s.mu.Lock()
	s.server = server
	s.mu.Unlock()
}

func (s *MCPLogSink) Core() zapcore.Core {
	return &mcpLogCore{sink: s}
}

type mcpLogCore struct {
	sink   *MCPLogSink
	fields []zapcore.Field
}

func (c *mcpLogCore) Enabled(level zapcore.Level) bool {
	return level >= c.sink.minLevel
}

func (c *mcpLogCore) With(fields []zapcore.Field) zapcore.Core {
	if len(fields) == 0 {
		return c
	}
	combined := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	combined = append(combined, c.fields...)
	combined = append(combined, fields...)
	return &mcpLogCore{sink: c.sink, fields: combined}
}

func (c *mcpLogCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

func (c *mcpLogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	server := c.currentServer()
	if server == nil {
		return nil
	}

	data := c.buildData(entry, fields)
	params := &mcp.LoggingMessageParams{
		Logger: c.loggerName(entry.LoggerName),
		Level:  mapZapLevel(entry.Level),
		Data:   data,
	}

	ctx := context.Background()
	for session := range server.Sessions() {
		_ = session.Log(ctx, params)
	}
	return nil
}

func (c *mcpLogCore) Sync() error {
	return nil
}

func (c *mcpLogCore) currentServer() *mcp.Server {
	c.sink.mu.RLock()
	defer c.sink.mu.RUnlock()
	return c.sink.server
}

func (c *mcpLogCore) loggerName(entryName string) string {
	if entryName != "" {
		return entryName
	}
	return c.sink.loggerName
}

func (c *mcpLogCore) buildData(entry zapcore.Entry, fields []zapcore.Field) map[string]any {
	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range c.fields {
		field.AddTo(encoder)
	}
	for _, field := range fields {
		field.AddTo(encoder)
	}

	data := map[string]any{
		"message":   entry.Message,
		"timestamp": entry.Time.UTC().Format(time.RFC3339Nano),
	}
	if entry.LoggerName != "" {
		data["logger"] = entry.LoggerName
	}
	if len(encoder.Fields) > 0 {
		data["fields"] = encoder.Fields
	}
	return data
}

func mapZapLevel(level zapcore.Level) mcp.LoggingLevel {
	switch level {
	case zapcore.DebugLevel:
		return "debug"
	case zapcore.InfoLevel:
		return "info"
	case zapcore.WarnLevel:
		return "warning"
	case zapcore.ErrorLevel:
		return "error"
	case zapcore.DPanicLevel:
		return "critical"
	case zapcore.PanicLevel:
		return "alert"
	case zapcore.FatalLevel:
		return "emergency"
	default:
		return "info"
	}
}

var _ zapcore.Core = (*mcpLogCore)(nil)
