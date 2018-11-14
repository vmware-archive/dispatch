///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

const (
	followPeriod time.Duration = 2 * time.Second
)

var (
	getRunLong = i18n.T(`Get run(s).`)

	getRunExample = i18n.T(`
# Get all runs
dispatch get runs

# Get runs for a specific function
dispatch get runs example-function

# Get a specific run
dispatch get runs example-function f98d0a7f-0c1d-4020-a488-cabc501b08e0

# Follow runs for a specific function
dispatch get runs example-function --follow

# Get runs sorted by finished time
dispatch get runs --by finished
`)

	followRuns = false
	last       = false
	sortBy     = ""
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
			c := functionManagerClient()
			if len(args) == 2 {
				opts := client.FunctionOpts{
					FunctionName: &args[0],
					RunName:      &args[1],
				}
				err = getFunctionRun(out, errOut, cmd, opts, c)
			} else if len(args) == 1 {
				opts := client.FunctionOpts{
					FunctionName: &args[0],
				}
				err = getRuns(out, errOut, cmd, opts, c)
			} else {
				opts := client.FunctionOpts{}
				err = getRuns(out, errOut, cmd, opts, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	cmd.Flags().BoolVarP(&followRuns, "follow", "f", false, "follow function runs, default: false")
	cmd.Flags().BoolVar(&last, "last", false, "get last executed run, default: false")
	cmd.Flags().StringVar(&sortBy, "by", "started", "sort runs by [function|status|started|finished]. default: started")
	return cmd
}

func getFunctionRun(out, errOut io.Writer, cmd *cobra.Command, opts client.FunctionOpts, c client.FunctionsClient) error {
	since := time.Now()
	resp, err := c.GetFunctionRun(context.TODO(), "", opts)

	if err != nil {
		return err
	}
	if err = formatRunOutput(out, false, true, []v1.Run{*resp}); err != nil {
		return err
	}
	if followRuns {
		opts.Since = since
		if err = followFilteredRuns(out, c, opts); err != nil {
			return err
		}
	}
	return nil
}

func getRuns(out, errOut io.Writer, cmd *cobra.Command, opts client.FunctionOpts, c client.FunctionsClient) error {
	since := time.Now()
	resp, err := c.ListRuns(context.TODO(), "", opts)

	if err != nil {
		return err
	}
	if last && len(resp) > 0 {
		lastRun := resp[0]
		for _, run := range resp {
			if run.ExecutedTime > lastRun.ExecutedTime {
				lastRun = run
			}
		}
		resp = []v1.Run{lastRun}
	}
	if err = formatRunOutput(out, true, true, resp); err != nil {
		return err
	}
	if followRuns {
		opts.Since = since
		if err = followFilteredRuns(out, c, opts); err != nil {
			return err
		}
	}

	return nil
}

func followFilteredRuns(out io.Writer, c client.FunctionsClient, opts client.FunctionOpts) error {
	before := opts.Since

	followTicker := time.NewTicker(followPeriod)
	defer followTicker.Stop()

	signals := make(chan os.Signal, 1)
	defer signal.Stop(signals)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-signals:
			signal.Stop(signals)
			followTicker.Stop()
			os.Exit(1)
		case t := <-followTicker.C:
			var resp []v1.Run
			var err error
			opts.Since = before
			before = t

			if opts.FunctionName == nil || opts.RunName == nil {
				resp, err = c.ListRuns(context.TODO(), "", opts)
			} else {
				var run *v1.Run
				run, err = c.GetFunctionRun(context.TODO(), "", opts)
				if _, ok := err.(*client.ErrorNotFound); ok {
					continue
				}
				if run != nil {
					resp = []v1.Run{*run}
				}
			}

			if err != nil {
				return err
			}
			if len(resp) > 0 {
				if err = formatRunOutput(out, true, false, resp); err != nil {
					return err
				}
			}
		}
	}
}

func formatRunOutput(out io.Writer, list bool, header bool, runs []v1.Run) error {

	var sortFuncs = map[string]func(i, j int) bool{
		"function": func(i, j int) bool {
			return runs[i].FunctionName < runs[j].FunctionName
		},
		"status": func(i, j int) bool {
			return runs[i].Status < runs[j].Status
		},
		"finished": func(i, j int) bool {
			return runs[i].FinishedTime > runs[j].FinishedTime
		},
		"started": func(i, j int) bool {
			return runs[i].ExecutedTime > runs[j].ExecutedTime
		},
	}

	if _, ok := sortFuncs[sortBy]; ok {
		sort.Slice(runs, sortFuncs[sortBy])
	}

	if w, err := formatOutput(out, list, runs); w {
		return err
	}
	table := tablewriter.NewWriter(out)
	if header {
		table.SetHeader([]string{"ID", "Function", "Status", "Started", "Finished"})
	}
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
