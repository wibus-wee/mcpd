package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"mcpd/internal/domain"
	"mcpd/internal/infra/telemetry"
)

type StdioTransport struct {
	logger *zap.Logger
}

type StdioTransportOptions struct {
	Logger *zap.Logger
}

func NewStdioTransport(opts StdioTransportOptions) *StdioTransport {
	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StdioTransport{logger: logger}
}

type processCleanup func()

func (t *StdioTransport) Start(ctx context.Context, spec domain.ServerSpec) (domain.Conn, domain.StopFn, error) {
	if len(spec.Cmd) == 0 {
		return nil, nil, errors.New("cmd is required for stdio transport")
	}

	cmd := exec.CommandContext(ctx, spec.Cmd[0], spec.Cmd[1:]...)
	if spec.Cwd != "" {
		cmd.Dir = spec.Cwd
	}
	cmd.Env = append(os.Environ(), formatEnv(spec.Env)...)
	groupCleanup := setupProcessHandling(cmd)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stderr pipe: %w", err)
	}
	downstreamLogger := t.logger.With(
		zap.String(telemetry.FieldLogSource, telemetry.LogSourceDownstream),
		telemetry.ServerTypeField(spec.Name),
		zap.String(telemetry.FieldLogStream, "stderr"),
	)
	go mirrorStderr(stderr, downstreamLogger)

	transport := &mcp.CommandTransport{
		Command: cmd,
	}

	mcpConn, err := transport.Connect(ctx)
	if err != nil {
		_ = stderr.Close()
		return nil, nil, fmt.Errorf("connect stdio: %w", err)
	}

	conn := &mcpConnAdapter{conn: mcpConn}
	stop := func(stopCtx context.Context) error {
		_ = mcpConn.Close()
		if groupCleanup != nil {
			groupCleanup()
		}
		return nil
	}

	return conn, stop, nil
}

type mcpConnAdapter struct {
	conn mcp.Connection
}

func (a *mcpConnAdapter) Send(ctx context.Context, msg json.RawMessage) error {
	if len(msg) == 0 {
		return errors.New("message is empty")
	}
	decoded, err := jsonrpc.DecodeMessage(msg)
	if err != nil {
		return fmt.Errorf("decode message: %w", err)
	}
	return a.conn.Write(ctx, decoded)
}

func (a *mcpConnAdapter) Recv(ctx context.Context) (json.RawMessage, error) {
	msg, err := a.conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	raw, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return nil, fmt.Errorf("encode message: %w", err)
	}
	return json.RawMessage(raw), nil
}

func (a *mcpConnAdapter) Close() error {
	return a.conn.Close()
}

func mirrorStderr(reader io.Reader, logger *zap.Logger) {
	buf := bufio.NewReader(reader)
	for {
		line, err := buf.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\r\n")
			if line != "" {
				logger.Info(line)
			}
		}
		if err != nil {
			return
		}
	}
}

func formatEnv(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}
