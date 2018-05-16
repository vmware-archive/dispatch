///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/client/runner"
)

var (
	getRunLong = i18n.T(`Get run(s).`)

	// TODO: add examples
	getRunExample = i18n.T(``)
)

// NewCmdGetRun creates command responsible for getting runs.
func NewCmdGetRun(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run [FUNCTION_NAME [RUN_ID]]",
		Short:   i18n.T("Get run(s)"),
		Long:    getRunLong,
		Example: getRunExample,
		Args:    cobra.RangeArgs(0, 2),
		Aliases: []string{"runs"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) == 2 {
				err = getFunctionRun(out, errOut, cmd, args)
			} else if len(args) == 1 {
				err = getFunctionRuns(out, errOut, cmd, args)
			} else {
				err = getRuns(out, errOut, cmd, args)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

func getFunctionRun(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := functionManagerClient()
	params := &fnrunner.GetRunParams{
		FunctionName: &args[0],
		RunName:      strfmt.UUID(args[1]),
		Context:      context.Background(),
		Tags:         []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.Runner.GetRun(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatRunOutput(out, false, []*v1.Run{resp.Payload})
}

func getRuns(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := functionManagerClient()
	params := &fnrunner.GetRunsParams{
		Context: context.Background(),
		Tags:    []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.Runner.GetRuns(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatRunOutput(out, true, resp.Payload)
}

func getFunctionRuns(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := functionManagerClient()
	params := &fnrunner.GetRunsParams{
		Context: context.Background(),
		Tags:    []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	if len(args) > 0 {
		params.FunctionName = &args[0]
	}
	resp, err := client.Runner.GetRuns(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatRunOutput(out, true, resp.Payload)
}

func formatRunOutput(out io.Writer, list bool, runs []*v1.Run) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(runs)
		}
		return encoder.Encode(runs[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"ID", "Function", "Status", "Started", "Finished"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, run := range runs {
		table.Append([]string{
			run.Name.String(),
			run.FunctionName,
			string(run.Status),
			time.Unix(run.ExecutedTime, 0).Local().Format(time.UnixDate),
			time.Unix(run.FinishedTime, 0).Local().Format(time.UnixDate),
		})
	}
	table.Render()
	return nil
}
