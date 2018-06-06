///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/vmware/dispatch/pkg/api/v1"
	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	drivers "github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
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

// CallDeleteEventDriver makes the API call to delete an event driver
func CallDeleteEventDriver(i interface{}) error {
	client := eventManagerClient()
	driverModel := i.(*v1.EventDriver)
	params := &drivers.DeleteDriverParams{
		Context:    context.Background(),
		DriverName: *driverModel.Name,
		Tags:       []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	deleted, err := client.Drivers.DeleteDriver(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*driverModel = *deleted.Payload
	return nil
}

func deleteEventDriver(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	driverModel := v1.EventDriver{
		Name: &args[0],
	}
	err := CallDeleteEventDriver(&driverModel)
	if err != nil {
		return err
	}
	return formatDeleteEventDriverOutput(out, false, []*v1.EventDriver{&driverModel})
}

func formatDeleteEventDriverOutput(out io.Writer, list bool, drivers []*v1.EventDriver) error {
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
