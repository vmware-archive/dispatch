///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"encoding/json"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	getEventDriverTypeLong = i18n.T(
		`Get dispatch event driver type.`)
	// TODO: add examples
	getEventDriverTypeExample     = i18n.T(``)
	getEventDriverTypeShowBuiltIn = false
)

// NewCmdGetEventDriverType gets command responsible for retrieving Dispatch event driver type.
func NewCmdGetEventDriverType(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eventdrivertype [DRIVER_TYPE_NAME] [--show-builtin]",
		Short:   i18n.T("Get event driver type"),
		Long:    getEventDriverTypeLong,
		Example: getEventDriverTypeExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"eventdrivertypes", "event-driver-type", "event-driver-types", "eventdriver-types", "eventdriver-type"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := eventManagerClient()
			if len(args) == 1 {
				err = getEventDriverType(out, errOut, cmd, args, c)
			} else {
				err = getEventDriverTypes(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	cmd.Flags().BoolVar(&getEventDriverTypeShowBuiltIn, "show-builtin", false, "Include built-in driver types in result")
	return cmd
}

func getEventDriverTypes(out, errOut io.Writer, cmd *cobra.Command, c client.EventsClient) error {

	get, err := c.ListEventDriverTypes(context.TODO(), "")
	if err != nil {
		return err
	}
	filtered := get
	if !getEventDriverTypeShowBuiltIn {
		// TODO: filter this server-side
		filtered = []v1.EventDriverType{}
		for _, t := range get {
			if !*t.BuiltIn {
				filtered = append(filtered, t)
			}
		}
	}

	return formatEventDriverTypeOutput(out, true, filtered)
}

func getEventDriverType(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EventsClient) error {

	driverTypeName := args[0]

	get, err := c.GetEventDriverType(context.TODO(), "", driverTypeName)
	if err != nil {
		return err
	}
	if !getEventDriverTypeShowBuiltIn && *get.BuiltIn {
		formatEventDriverTypeOutput(out, false, []v1.EventDriverType{})
	}
	return formatEventDriverTypeOutput(out, false, []v1.EventDriverType{*get})
}

func formatEventDriverTypeOutput(out io.Writer, list bool, driverTypes []v1.EventDriverType) error {

	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(driverTypes)
		}
		return encoder.Encode(driverTypes[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Image"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("-")
	table.SetRowLine(true)
	for _, d := range driverTypes {
		table.Append([]string{*d.Name, *d.Image})
	}
	table.Render()
	return nil
}
