package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type payloadFlags struct {
	args string
	file string
}

func bindPayloadFlags(cmd *cobra.Command, label string) *payloadFlags {
	flags := &payloadFlags{}
	cmd.Flags().StringVar(&flags.args, "args", "", fmt.Sprintf("%s arguments as JSON", label))
	cmd.Flags().StringVar(&flags.file, "args-file", "", fmt.Sprintf("path to JSON file with %s arguments", label))
	return flags
}

func bindRoutingKeyFlag(cmd *cobra.Command) *string {
	var routingKey string
	cmd.Flags().StringVar(&routingKey, "routing-key", "", "routing key")
	return &routingKey
}

func bindCursorFlag(cmd *cobra.Command, help string) *string {
	var cursor string
	cmd.Flags().StringVar(&cursor, "cursor", "", help)
	return &cursor
}

func bindLimitFlag(cmd *cobra.Command, help string) *int32 {
	var limit int32
	cmd.Flags().Int32Var(&limit, "limit", 0, help)
	return &limit
}

func bindLastETagFlag(cmd *cobra.Command) *string {
	var lastETag string
	cmd.Flags().StringVar(&lastETag, "last-etag", "", "resume from etag")
	return &lastETag
}

func (flags payloadFlags) loadPayload() ([]byte, error) {
	return loadJSONPayload(flags.args, flags.file)
}
