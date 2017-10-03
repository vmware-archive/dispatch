///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"fmt"
	"io"
	"time"

	"golang.org/x/net/context"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client/store"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
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
			var err error
			if len(args) > 0 {
				err = getFunction(out, errOut, cmd, args)
			} else {
				err = getFunctions(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getFunction(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := functionManagerClient()
	params := &fnstore.GetFunctionByNameParams{
		FunctionName: args[0],
		Context:      context.Background(),
	}

	response, err := client.Store.GetFunctionByName(params)
	if err != nil {
		fmt.Fprintf(errOut, "Error when retreving function %s\n", args[0])
		return err
	}
	function := response.Payload
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Attribute", "value"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	table.Append([]string{"Name", *function.Name})
	table.Append([]string{"Image", *function.Image})
	table.Append([]string{"State", string(function.State)})
	table.Append([]string{"Time created", time.Unix(function.CreatedTime, 0).Local().Format(time.UnixDate)})
	table.Append([]string{"Time modified", time.Unix(function.ModifiedTime, 0).Local().Format(time.UnixDate)})
	table.Render()
	return nil
}

func getFunctions(out, errOut io.Writer, cmd *cobra.Command) error {
	client := functionManagerClient()
	params := &fnstore.GetFunctionsParams{
		Context: context.Background(),
	}
	functions, err := client.Store.GetFunctions(params)
	if err != nil {
		fmt.Fprintf(errOut, "Error when retreiving functions\n")
		return err
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Image", "State", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")

	for _, function := range functions.Payload {
		table.Append([]string{*function.Name, *function.Image, string(function.State), time.Unix(function.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
