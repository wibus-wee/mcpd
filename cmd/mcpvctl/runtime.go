package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newRuntimeCmd(opts *cliOptions) *cobra.Command {
	var lastETag *string
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Runtime status",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "watch",
		Short: "Watch runtime status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := signalAwareContext(cmd.Context())
			defer cancel()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				stream, err := client.WatchRuntimeStatus(ctx, &controlv1.WatchRuntimeStatusRequest{Caller: caller, LastEtag: strings.TrimSpace(*lastETag)})
				if err != nil {
					return err
				}
				return watchStream(stream.Recv, func(snapshot *controlv1.RuntimeStatusSnapshot) error {
					if opts.jsonOutput {
						return writeJSON(snapshot)
					}
					fmt.Printf("runtime snapshot etag=%s servers=%d\n", snapshot.GetEtag(), len(snapshot.GetStatuses()))
					return nil
				})
			})
		},
	})
	lastETag = bindLastETagFlag(cmd)
	return cmd
}
