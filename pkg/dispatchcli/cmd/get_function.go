///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	getFunctionLong = i18n.T(`Get function(s).`)

	// TODO: add examples
	getFunctionExample = i18n.T(``)
)

// NewCmdGetFunction creates command responsible for getting functions.
func NewCmdGetFunction(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "function [FUNCTION_NAME]",
		Short:   i18n.T("Get function(s)"),
		Long:    getFunctionLong,
		Example: getFunctionExample,
		Args:    cobra.RangeArgs(0, 1),
		Aliases: []string{"functions"},
		Run: func(cmd *cobra.Command, args []string) {
			c := functionsClient()
			var err error
			if len(args) > 0 {
				err = getFunction(out, errOut, cmd, args, c)
			} else {
				err = getFunctions(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}

	return cmd
}

func getFunction(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.FunctionsClient) error {
	functionName := args[0]

	resp, err := c.GetFunction(context.TODO(), dispatchConfig.Organization, functionName)
	if err != nil {
		return err
	}

	return formatFunctionOutput(out, false, []v1.Function{*resp})
}

func getFunctions(out, errOut io.Writer, cmd *cobra.Command, c client.FunctionsClient) error {
	resp, err := c.ListFunctions(context.TODO(), dispatchConfig.Organization)
	if err != nil {
		return err
	}
	return formatFunctionOutput(out, true, resp)
}

func formatFunctionOutput(out io.Writer, list bool, functions []v1.Function) error {
	if w, err := formatOutput(out, list, functions); w {
		return err
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "FunctionImage", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, function := range functions {
		table.Append([]string{function.Meta.Name, function.FunctionImageURL, string(function.Status), time.Unix(function.Meta.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
