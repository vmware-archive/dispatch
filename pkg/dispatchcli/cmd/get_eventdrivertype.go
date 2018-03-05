///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	client "github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
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
			if len(args) == 1 {
				err = getEventDriverType(out, errOut, cmd, args)
			} else {
				err = getEventDriverTypes(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	cmd.Flags().BoolVar(&getEventDriverTypeShowBuiltIn, "show-builtin", false, "Include built-in driver types in result")
	return cmd
}

func getEventDriverTypes(out, errOut io.Writer, cmd *cobra.Command) error {

	params := &client.GetDriverTypesParams{
		Context: context.Background(),
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	get, err := eventManagerClient().Drivers.GetDriverTypes(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	filtered := get.Payload
	if !getEventDriverTypeShowBuiltIn {
		// TODO: filter this server-side
		filtered = []*models.DriverType{}
		for _, t := range get.Payload {
			if !*t.BuiltIn {
				filtered = append(filtered, t)
			}
		}
	}

	return formatEventDriverTypeOutput(out, true, filtered)
}

func getEventDriverType(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	driverTypeName := args[0]
	params := &client.GetDriverTypeParams{
		DriverTypeName: driverTypeName,
		Context:        context.Background(),
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	get, err := eventManagerClient().Drivers.GetDriverType(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if !getEventDriverTypeShowBuiltIn && *get.Payload.BuiltIn {
		formatEventDriverTypeOutput(out, false, []*models.DriverType{})
	}
	return formatEventDriverTypeOutput(out, false, []*models.DriverType{get.Payload})
}

func formatEventDriverTypeOutput(out io.Writer, list bool, driverTypes []*models.DriverType) error {

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
