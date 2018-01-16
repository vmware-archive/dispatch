///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/events"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

var (
	emitLong = i18n.T(`Emit an event.`)

	// TODO: Add examples
	emitExample = i18n.T(``)

	emitWait    = false
	emitPayload = "{}"
)

// NewCmdEmit creates a command to emit a dispatch event.
func NewCmdEmit(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "emit [--payload JSON] TOPIC_NAME",
		Short:   i18n.T("Emit a dispatch event"),
		Long:    emitLong,
		Example: emitExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := runEmit(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&emitPayload, "payload", "{}", "Event payload JSON object")
	return cmd
}

func runEmit(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	eventTopic := args[0]
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(emitPayload), &payload)
	if err != nil {
		fmt.Fprintf(errOut, "Error when parsing event payload %s\n", emitPayload)
		return err
	}
	emission := &models.Emission{
		Topic:   &eventTopic,
		Payload: payload,
	}

	params := &events.EmitEventParams{
		Context: context.Background(),
		Body:    emission,
	}
	client := eventManagerClient()
	_, err = client.Events.EmitEvent(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	fmt.Fprintln(out, "event emitted")
	return nil
}
