package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newToolsCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Tool operations",
	}
	cmd.AddCommand(
		newToolsListCmd(opts),
		newToolsWatchCmd(opts),
		newToolsCallCmd(opts),
		newToolsCallTaskCmd(opts),
	)
	return cmd
}

func newToolsListCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.ListTools(ctx, &controlv1.ListToolsRequest{Caller: caller})
				if err != nil {
					return err
				}
				return printToolsSnapshot(resp.GetSnapshot(), opts.jsonOutput)
			})
		},
	}
	return cmd
}

func newToolsWatchCmd(opts *cliOptions) *cobra.Command {
	var lastETag *string
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch tool snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := signalAwareContext(cmd.Context())
			defer cancel()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				stream, err := client.WatchTools(ctx, &controlv1.WatchToolsRequest{Caller: caller, LastEtag: strings.TrimSpace(*lastETag)})
				if err != nil {
					return err
				}
				return watchStream(stream.Recv, func(snapshot *controlv1.ToolsSnapshot) error {
					return printToolsSnapshot(snapshot, opts.jsonOutput)
				})
			})
		},
	}
	lastETag = bindLastETagFlag(cmd)
	return cmd
}

func newToolsCallCmd(opts *cliOptions) *cobra.Command {
	var payloads *payloadFlags
	var routingKey *string
	cmd := &cobra.Command{
		Use:   "call <tool-name>",
		Short: "Call a tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := strings.TrimSpace(args[0])
			payload, err := payloads.loadPayload()
			if err != nil {
				return err
			}
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.CallTool(ctx, &controlv1.CallToolRequest{
					Caller:        caller,
					Name:          name,
					ArgumentsJson: payload,
					RoutingKey:    strings.TrimSpace(*routingKey),
				})
				if err != nil {
					return err
				}
				return printResultPayload("result", resp.GetResultJson(), opts.jsonOutput)
			})
		},
	}
	payloads = bindPayloadFlags(cmd, "tool")
	routingKey = bindRoutingKeyFlag(cmd)
	return cmd
}

func newToolsCallTaskCmd(opts *cliOptions) *cobra.Command {
	var payloads *payloadFlags
	var routingKey *string
	var ttlMs int64
	var pollMs int64
	cmd := &cobra.Command{
		Use:   "call-task <tool-name>",
		Short: "Call a tool asynchronously (task)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := strings.TrimSpace(args[0])
			payload, err := payloads.loadPayload()
			if err != nil {
				return err
			}
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.CallToolTask(ctx, &controlv1.CallToolTaskRequest{
					Caller:         caller,
					Name:           name,
					ArgumentsJson:  payload,
					RoutingKey:     strings.TrimSpace(*routingKey),
					TtlMs:          ttlMs,
					PollIntervalMs: pollMs,
				})
				if err != nil {
					return err
				}
				return printTask(resp.GetTask(), opts.jsonOutput)
			})
		},
	}
	payloads = bindPayloadFlags(cmd, "tool")
	routingKey = bindRoutingKeyFlag(cmd)
	cmd.Flags().Int64Var(&ttlMs, "ttl-ms", 0, "task TTL in milliseconds")
	cmd.Flags().Int64Var(&pollMs, "poll-interval-ms", 0, "task poll interval in milliseconds")
	return cmd
}
