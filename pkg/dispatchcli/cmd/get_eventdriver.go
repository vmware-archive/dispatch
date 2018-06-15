///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	getEventDriverLong = i18n.T(
		`Get dispatch event driver.`)
	// TODO: add examples
	getEventDriverExample = i18n.T(``)
)

// NewCmdGetEventDriver gets command responsible for retrieving Dispatch event driver.
func NewCmdGetEventDriver(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eventdriver [DRIVER_NAME]",
		Short:   i18n.T("Get Event Driver"),
		Long:    getEventDriverLong,
		Example: getEventDriverExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"eventdrivers", "event-driver", "event-drivers"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := eventManagerClient()
			if len(args) == 1 {
				err = getEventDriver(out, errOut, cmd, args, c)
			} else {
				err = getEventDrivers(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

func getEventDrivers(out, errOut io.Writer, cmd *cobra.Command, c client.EventsClient) error {

	get, err := c.ListEventDrivers(context.TODO(), "")
	if err != nil {
		return err
	}
	return formatEventDriverOutput(out, true, get)
}

func getEventDriver(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EventsClient) error {

	driverName := args[0]

	get, err := c.GetEventDriver(context.TODO(), "", driverName)
	if err != nil {
		return err
	}

	return formatEventDriverOutput(out, false, []v1.EventDriver{*get})
}

func formatEventDriverOutput(out io.Writer, list bool, drivers []v1.EventDriver) error {

	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(drivers)
		}
		return encoder.Encode(drivers[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Type", "Status", "Secrets", "Config", "Reason"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("-")
	table.SetRowLine(true)
	for _, d := range drivers {
		var configs []string
		for _, c := range d.Config {
			if c.Value == "" {
				configs = append(configs, fmt.Sprintf("%s", c.Key))
			} else {
				configs = append(configs, fmt.Sprintf("%s=%s", c.Key, c.Value))
			}
		}

		table.Append([]string{
			*d.Name, *d.Type, fmt.Sprintf("%s", d.Status),
			strings.Join(d.Secrets, ","),
			strings.Join(configs, "\n"),
			strings.Join(d.Reason, ", "),
		})
	}
	table.Render()
	return nil
}
