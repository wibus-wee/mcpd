package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	controlv1 "mcpv/pkg/api/control/v1"
)

func newTasksCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Task operations",
	}
	cmd.AddCommand(
		newTasksListCmd(opts),
		newTasksGetCmd(opts),
		newTasksResultCmd(opts),
		newTasksCancelCmd(opts),
	)
	return cmd
}

func newTasksListCmd(opts *cliOptions) *cobra.Command {
	var cursor *string
	var limit *int32
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.TasksList(ctx, &controlv1.TasksListRequest{
					Caller: caller,
					Cursor: strings.TrimSpace(*cursor),
					Limit:  *limit,
				})
				if err != nil {
					return err
				}
				return printTasksList(resp, opts.jsonOutput)
			})
		},
	}
	cursor = bindCursorFlag(cmd, "pagination cursor")
	limit = bindLimitFlag(cmd, "page size")
	return cmd
}

func newTasksGetCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <task-id>",
		Short: "Get task status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			id := strings.TrimSpace(args[0])
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.TasksGet(ctx, &controlv1.TasksGetRequest{Caller: caller, TaskId: id})
				if err != nil {
					return err
				}
				return printTask(resp.GetTask(), opts.jsonOutput)
			})
		},
	}
	return cmd
}

func newTasksResultCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "result <task-id>",
		Short: "Get task result",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			id := strings.TrimSpace(args[0])
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				resp, err := client.TasksResult(ctx, &controlv1.TasksResultRequest{Caller: caller, TaskId: id})
				if err != nil {
					return err
				}
				return printTaskResult(resp.GetResult(), opts.jsonOutput)
			})
		},
	}
	return cmd
}

func newTasksCancelCmd(opts *cliOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <task-id>",
		Short: "Cancel a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			id := strings.TrimSpace(args[0])
			return withSession(ctx, opts, func(ctx context.Context, client controlv1.ControlPlaneServiceClient, caller string) error {
				_, err := client.TasksCancel(ctx, &controlv1.TasksCancelRequest{Caller: caller, TaskId: id})
				if err != nil {
					return err
				}
				if opts.jsonOutput {
					return writeJSON(map[string]string{"status": "ok"})
				}
				fmt.Printf("Task %s canceled\n", id)
				return nil
			})
		},
	}
	return cmd
}
