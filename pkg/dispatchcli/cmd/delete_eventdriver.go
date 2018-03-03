///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	client "github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

var (
	deleteEventDriverLong = i18n.T(`Delete event driver`)

	// TODO: add examples
	deleteEventDriverExample = i18n.T(``)
)

// NewCmdDeleteEventDriver creates command responsible for deleting EventDriver.
func NewCmdDeleteEventDriver(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eventdriver DRIVER_NAME",
		Short:   i18n.T("Delete EventDriver"),
		Long:    deleteEventDriverLong,
		Example: deleteEventDriverExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"eventdrivers", "event-driver", "event-drivers"},
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteEventDriver(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

func deleteEventDriver(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	params := &client.DeleteDriverParams{
		Context:    context.Background(),
		DriverName: args[0],
		Tags:       []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := eventManagerClient().Drivers.DeleteDriver(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatDeleteEventDriverOutput(out, false, []*models.Driver{resp.Payload})
}

func formatDeleteEventDriverOutput(out io.Writer, list bool, drivers []*models.Driver) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(drivers)
		}
		return encoder.Encode(drivers[0])
	}
	for _, d := range drivers {
		_, err := fmt.Fprintf(out, "Deleted drivers: %s\n", *d.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
