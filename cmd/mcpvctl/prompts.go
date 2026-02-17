package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newPromptsCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompts",
		Short: "Prompt operations",
	}
	cmd.AddCommand(
		newPromptsListCmd(opts),
		newPromptsWatchCmd(opts),
		newPromptsGetCmd(opts),
	)
	return cmd
}

func newPromptsListCmd(opts *cliOptions) *cobra.Command {
	var cursor *string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prompts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.ListPrompts(ctx, &controlv1.ListPromptsRequest{Caller: caller, Cursor: strings.TrimSpace(*cursor)})
				if err != nil {
					return err
				}
				return printPromptsSnapshot(resp.GetSnapshot(), resp.GetNextCursor(), opts.jsonOutput)
			})
		},
	}
	cursor = bindCursorFlag(cmd, "pagination cursor")
	return cmd
}

func newPromptsWatchCmd(opts *cliOptions) *cobra.Command {
	var lastETag *string
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch prompt snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := signalAwareContext(cmd.Context())
			defer cancel()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				stream, err := client.WatchPrompts(ctx, &controlv1.WatchPromptsRequest{Caller: caller, LastEtag: strings.TrimSpace(*lastETag)})
				if err != nil {
					return err
				}
				return watchStream(stream.Recv, func(snapshot *controlv1.PromptsSnapshot) error {
					return printPromptsSnapshot(snapshot, "", opts.jsonOutput)
				})
			})
		},
	}
	lastETag = bindLastETagFlag(cmd)
	return cmd
}

func newPromptsGetCmd(opts *cliOptions) *cobra.Command {
	var payloads *payloadFlags
	cmd := &cobra.Command{
		Use:   "get <prompt-name>",
		Short: "Get a prompt",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := strings.TrimSpace(args[0])
			payload, err := payloads.loadPayload()
			if err != nil {
				return err
			}
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.GetPrompt(ctx, &controlv1.GetPromptRequest{
					Caller:        caller,
					Name:          name,
					ArgumentsJson: payload,
				})
				if err != nil {
					return err
				}
				return printResultPayload("result", resp.GetResultJson(), opts.jsonOutput)
			})
		},
	}
	payloads = bindPayloadFlags(cmd, "prompt")
	return cmd
}
