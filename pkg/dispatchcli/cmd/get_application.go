///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/application-manager/gen/client/application"
	"github.com/vmware/dispatch/pkg/application-manager/gen/models"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	getApplicationsLong = i18n.T(`Get applications.`)

	// TODO: add examples
	getApplicationsExample = i18n.T(``)
)

// NewCmdGetApplication creates command responsible for getting applications.
func NewCmdGetApplication(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application [APPLICATION]",
		Short:   i18n.T("Get applications"),
		Long:    getApplicationsLong,
		Example: getApplicationsExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"app", "apps", "applications"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) > 0 {
				err = getApplication(out, errOut, cmd, args)
			} else {
				err = getApplications(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getApplication(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := applicationManagerClient()
	params := &application.GetAppParams{
		Context:     context.Background(),
		Application: args[0],
	}

	resp, err := client.Application.GetApp(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatApplicationOutput(out, false, []*models.Application{resp.Payload})
}

func getApplications(out, errOut io.Writer, cmd *cobra.Command) error {
	client := applicationManagerClient()
	params := &application.GetAppsParams{
		Context: context.Background(),
	}
	resp, err := client.Application.GetApps(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatApplicationOutput(out, true, resp.Payload)
}

func formatApplicationOutput(out io.Writer, list bool, applications []*models.Application) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(applications)
		}
		return encoder.Encode(applications[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, app := range applications {
		table.Append([]string{*app.Name, string(app.Status), time.Unix(app.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
