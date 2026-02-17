package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newLogsCmd(opts *cliOptions) *cobra.Command {
	var minLevel string
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Stream logs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			level, err := parseLogLevel(minLevel)
			if err != nil {
				return err
			}
			ctx, cancel := signalAwareContext(cmd.Context())
			defer cancel()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				stream, err := client.StreamLogs(ctx, &controlv1.StreamLogsRequest{Caller: caller, MinLevel: level})
				if err != nil {
					return err
				}
				return watchStream(stream.Recv, func(entry *controlv1.LogEntry) error {
					return printLogEntry(entry, opts.jsonOutput)
				})
			})
		},
	}
	cmd.Flags().StringVar(&minLevel, "min-level", "info", "minimum log level")
	return cmd
}

func parseLogLevel(value string) (controlv1.LogLevel, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info":
		return controlv1.LogLevel_LOG_LEVEL_INFO, nil
	case "debug":
		return controlv1.LogLevel_LOG_LEVEL_DEBUG, nil
	case "notice":
		return controlv1.LogLevel_LOG_LEVEL_NOTICE, nil
	case "warn", "warning":
		return controlv1.LogLevel_LOG_LEVEL_WARNING, nil
	case "error":
		return controlv1.LogLevel_LOG_LEVEL_ERROR, nil
	case "critical":
		return controlv1.LogLevel_LOG_LEVEL_CRITICAL, nil
	case "alert":
		return controlv1.LogLevel_LOG_LEVEL_ALERT, nil
	case "emergency":
		return controlv1.LogLevel_LOG_LEVEL_EMERGENCY, nil
	default:
		return controlv1.LogLevel_LOG_LEVEL_UNSPECIFIED, fmt.Errorf("unknown log level: %s", value)
	}
}
