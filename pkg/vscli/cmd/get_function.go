///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"io"
	"time"

	"golang.org/x/net/context"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	fnstore "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client/store"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/models"
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
	params := &fnstore.GetFunctionParams{
		FunctionName: args[0],
		Context:      context.Background(),
	}

	resp, err := client.Store.GetFunction(params)
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatFunctionOutput(out, false, []*models.Function{resp.Payload})
}

func getFunctions(out, errOut io.Writer, cmd *cobra.Command) error {
	client := functionManagerClient()
	params := &fnstore.GetFunctionsParams{
		Context: context.Background(),
	}
	resp, err := client.Store.GetFunctions(params)
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatFunctionOutput(out, true, resp.Payload)
}

func formatFunctionOutput(out io.Writer, list bool, functions []*models.Function) error {
	if vsConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(functions)
		}
		return encoder.Encode(functions[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Image", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, function := range functions {
		table.Append([]string{*function.Name, *function.Image, string(function.State), time.Unix(function.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
