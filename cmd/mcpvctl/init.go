package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newInitCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Server init status",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "watch",
		Short: "Watch server init status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := signalAwareContext(cmd.Context())
			defer cancel()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				stream, err := client.WatchServerInitStatus(ctx, &controlv1.WatchServerInitStatusRequest{Caller: caller})
				if err != nil {
					return err
				}
				return watchStream(stream.Recv, func(snapshot *controlv1.ServerInitStatusSnapshot) error {
					if opts.jsonOutput {
						return writeJSON(snapshot)
					}
					fmt.Printf("init snapshot generated_at=%d servers=%d\n", snapshot.GetGeneratedAtUnixNano(), len(snapshot.GetStatuses()))
					return nil
				})
			})
		},
	})
	return cmd
}
