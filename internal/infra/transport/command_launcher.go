package transport

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"

	"mcpd/internal/domain"
	"mcpd/internal/infra/telemetry"
)

type CommandLauncher struct {
	logger *zap.Logger
}

type CommandLauncherOptions struct {
	Logger *zap.Logger
}

func NewCommandLauncher(opts CommandLauncherOptions) *CommandLauncher {
	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CommandLauncher{logger: logger}
}

func (l *CommandLauncher) Start(ctx context.Context, specKey string, spec domain.ServerSpec) (domain.IOStreams, domain.StopFn, error) {
	if len(spec.Cmd) == 0 {
		return domain.IOStreams{}, nil, fmt.Errorf("%w: cmd is required for stdio launcher", domain.ErrInvalidCommand)
	}

	cmd := exec.CommandContext(ctx, spec.Cmd[0], spec.Cmd[1:]...)
	if spec.Cwd != "" {
		cmd.Dir = spec.Cwd
	}
	cmd.Env = append(os.Environ(), formatEnv(spec.Env)...)
	groupCleanup := setupProcessHandling(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return domain.IOStreams{}, nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return domain.IOStreams{}, nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return domain.IOStreams{}, nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return domain.IOStreams{}, nil, fmt.Errorf("start command: %w", classifyStartError(err))
	}

	downstreamLogger := l.logger.With(
		zap.String(telemetry.FieldLogSource, telemetry.LogSourceDownstream),
		telemetry.ServerTypeField(spec.Name),
		zap.String(telemetry.FieldLogStream, "stderr"),
	)
	go mirrorStderr(stderr, downstreamLogger)

	stop := func(stopCtx context.Context) error {
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		if groupCleanup != nil {
			groupCleanup()
		}
		return waitForProcess(stopCtx, cmd)
	}

	return domain.IOStreams{Reader: stdout, Writer: stdin}, stop, nil
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

func classifyStartError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s", domain.ErrExecutableNotFound, err.Error())
	}
	if errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("%w: %s", domain.ErrPermissionDenied, err.Error())
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		if errors.Is(pathErr.Err, exec.ErrNotFound) || errors.Is(pathErr.Err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", domain.ErrExecutableNotFound, err.Error())
		}
		if errors.Is(pathErr.Err, os.ErrPermission) {
			return fmt.Errorf("%w: %s", domain.ErrPermissionDenied, err.Error())
		}
	}
	return err
}
