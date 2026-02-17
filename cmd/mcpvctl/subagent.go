package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newSubAgentCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subagent",
		Short: "SubAgent operations",
	}
	cmd.AddCommand(
		newSubAgentEnabledCmd(opts),
		newSubAgentMCPCmd(opts),
		newSubAgentEvalCmd(opts),
	)
	return cmd
}

func newSubAgentEnabledCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enabled",
		Short: "Check SubAgent enablement for caller",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.IsSubAgentEnabled(ctx, &controlv1.IsSubAgentEnabledRequest{Caller: caller})
				if err != nil {
					return err
				}
				if opts.jsonOutput {
					return writeJSON(map[string]bool{"enabled": resp.GetEnabled()})
				}
				fmt.Printf("SubAgent enabled: %t\n", resp.GetEnabled())
				return nil
			})
		},
	}
	return cmd
}

func newSubAgentMCPCmd(opts *cliOptions) *cobra.Command {
	var sessionID string
	var forceRefresh bool
	cmd := &cobra.Command{
		Use:   "mcp <query>",
		Short: "Automatic MCP tool discovery",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			ctx := cmd.Context()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.AutomaticMCP(ctx, &controlv1.AutomaticMCPRequest{
					Caller:       caller,
					Query:        query,
					SessionId:    strings.TrimSpace(sessionID),
					ForceRefresh: forceRefresh,
				})
				if err != nil {
					return err
				}
				return printAutomaticMCP(resp, opts.jsonOutput)
			})
		},
	}
	cmd.Flags().StringVar(&sessionID, "session-id", "", "session identifier")
	cmd.Flags().BoolVar(&forceRefresh, "force-refresh", false, "force refresh")
	return cmd
}

func newSubAgentEvalCmd(opts *cliOptions) *cobra.Command {
	var payloads *payloadFlags
	var routingKey *string
	cmd := &cobra.Command{
		Use:   "eval <tool-name>",
		Short: "Automatic tool evaluation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			payload, err := payloads.loadPayload()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.AutomaticEval(ctx, &controlv1.AutomaticEvalRequest{
					Caller:        caller,
					ToolName:      name,
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
