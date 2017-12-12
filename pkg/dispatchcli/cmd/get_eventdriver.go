///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	client "github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

var (
	getEventDriverLong = i18n.T(
		`Get serverless event driver.`)
	// TODO: add examples
	getEventDriverExample = i18n.T(``)
)

// NewCmdGetEventDriver gets command responsible for retrieving Dispatch event driver.
func NewCmdGetEventDriver(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "event-driver [EventDriver_NAME]",
		Short:   i18n.T("Get EventDriver"),
		Long:    getEventDriverLong,
		Example: getEventDriverExample,
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) == 1 {
				err = getEventDriver(out, errOut, cmd, args)
			} else {
				err = getEventDrivers(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getEventDrivers(out, errOut io.Writer, cmd *cobra.Command) error {

	params := &client.GetDriversParams{
		Context: context.Background(),
	}
	get, err := eventManagerClient().Drivers.GetDrivers(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatEventDriverOutput(out, true, get.Payload)
}

func getEventDriver(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	driverName := args[0]
	params := &client.GetDriverParams{
		DriverName: driverName,
		Context:    context.Background(),
	}
	get, err := eventManagerClient().Drivers.GetDriver(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return formatEventDriverOutput(out, false, []*models.Driver{get.Payload})
}

func formatEventDriverOutput(out io.Writer, list bool, drivers []*models.Driver) error {

	if dispatchConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(drivers)
		}
		return encoder.Encode(drivers[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Type", "Status", "Configs"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("-")
	table.SetRowLine(true)
	for _, d := range drivers {
		configs := []string{}
		for _, c := range d.Config {
			configs = append(configs, fmt.Sprintf("%s=%s", c.Key, c.Value))
		}
		table.Append([]string{
			*d.Name, *d.Type, fmt.Sprintf("%s", d.Status),
			strings.Join(configs, "\n"),
		})
	}
	table.Render()
	return nil
}
