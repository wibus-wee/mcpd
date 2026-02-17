package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newInfoCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show core build info",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return withClient(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient) error {
				resp, err := client.GetInfo(ctx, &controlv1.GetInfoRequest{})
				if err != nil {
					return err
				}
				if opts.jsonOutput {
					return writeJSON(resp)
				}
				fmt.Printf("Name: %s\nVersion: %s\nBuild: %s\n", resp.GetName(), resp.GetVersion(), resp.GetBuild())
				return nil
			})
		},
	}
	return cmd
}

func newRegisterCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register caller with core",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			caller := resolveCaller(opts.caller)
			return withClient(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient) error {
				resp, err := client.RegisterCaller(ctx, &controlv1.RegisterCallerRequest{
					Caller: caller,
					Pid:    int64(os.Getpid()),
					Tags:   normalizeTags(opts.tags),
					Server: strings.TrimSpace(opts.server),
				})
				if err != nil {
					return err
				}
				if opts.jsonOutput {
					return writeJSON(map[string]string{"profile": resp.GetProfile()})
				}
				fmt.Printf("Registered caller %q (profile: %s)\n", caller, resp.GetProfile())
				return nil
			})
		},
	}
	return cmd
}

func newUnregisterCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister",
		Short: "Unregister caller from core",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			caller := resolveCaller(opts.caller)
			return withClient(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient) error {
				_, err := client.UnregisterCaller(ctx, &controlv1.UnregisterCallerRequest{Caller: caller})
				if err != nil {
					return err
				}
				if opts.jsonOutput {
					return writeJSON(map[string]string{"status": "ok"})
				}
				fmt.Printf("Unregistered caller %q\n", caller)
				return nil
			})
		},
	}
	return cmd
}
