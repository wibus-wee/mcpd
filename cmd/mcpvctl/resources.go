package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newResourcesCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Resource operations",
	}
	cmd.AddCommand(
		newResourcesListCmd(opts),
		newResourcesWatchCmd(opts),
		newResourcesReadCmd(opts),
	)
	return cmd
}

func newResourcesListCmd(opts *cliOptions) *cobra.Command {
	var cursor *string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.ListResources(ctx, &controlv1.ListResourcesRequest{Caller: caller, Cursor: strings.TrimSpace(*cursor)})
				if err != nil {
					return err
				}
				return printResourcesSnapshot(resp.GetSnapshot(), resp.GetNextCursor(), opts.jsonOutput)
			})
		},
	}
	cursor = bindCursorFlag(cmd, "pagination cursor")
	return cmd
}

func newResourcesWatchCmd(opts *cliOptions) *cobra.Command {
	var lastETag *string
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch resource snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := signalAwareContext(cmd.Context())
			defer cancel()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				stream, err := client.WatchResources(ctx, &controlv1.WatchResourcesRequest{Caller: caller, LastEtag: strings.TrimSpace(*lastETag)})
				if err != nil {
					return err
				}
				return watchStream(stream.Recv, func(snapshot *controlv1.ResourcesSnapshot) error {
					return printResourcesSnapshot(snapshot, "", opts.jsonOutput)
				})
			})
		},
	}
	lastETag = bindLastETagFlag(cmd)
	return cmd
}

func newResourcesReadCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <uri>",
		Short: "Read a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uri := strings.TrimSpace(args[0])
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.ReadResource(ctx, &controlv1.ReadResourceRequest{Caller: caller, Uri: uri})
				if err != nil {
					return err
				}
				return printResultPayload("result", resp.GetResultJson(), opts.jsonOutput)
			})
		},
	}
	return cmd
}
