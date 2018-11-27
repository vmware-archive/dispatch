///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	getAPILong = i18n.T(
		`Get dispatch function api.`)
	// TODO: add examples
	getAPIExample = i18n.T(``)

	functionName = ""
)

// NewCmdGetAPI gets command responsible for dispatch function api creation.
func NewCmdGetAPI(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api [API_NAME] [--func FUNC_NAME]",
		Short:   i18n.T("Get API"),
		Long:    getAPILong,
		Example: getAPIExample,
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := apiManagerClient()
			if len(args) == 1 {
				err = getAPI(out, errOut, cmd, args, c)
			} else {
				err = getAPIs(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	cmd.Flags().StringVarP(&functionName, "func", "f", "", "get all apis for specified function")
	return cmd
}

func getAPIs(out, errOut io.Writer, cmd *cobra.Command, c client.APIsClient) error {
	get, err := c.ListAPIs(context.TODO(), "")
	if err != nil {
		return err
	}
	return formatAPIOutput(out, true, get)
}

func getAPI(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.APIsClient) error {

	apiName := args[0]
	get, err := c.GetAPI(context.TODO(), "", apiName)
	if err != nil {
		return err
	}

	return formatAPIOutput(out, false, []v1.API{*get})
}

func formatAPIOutput(out io.Writer, list bool, apis []v1.API) error {

	if w, err := formatOutput(out, list, apis); w {
		return err
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Function", "Protocol", "Method", "Domain", "Path", "Auth", "Status", "Enabled"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("-")
	table.SetRowLine(true)
	for _, a := range apis {
		table.Append([]string{
			*a.Name, *a.Function,
			strings.Join(a.Protocols, "\n"), strings.Join(a.Methods, "\n"), strings.Join(a.Hosts, "\n"),
			strings.Join(addHostName(a.Uris, a.Protocols), "\n"), a.Authentication, string(a.Status), fmt.Sprintf("%t", a.Enabled),
		})
	}
	table.Render()
	return nil
}

func addHostName(uris []string, protocols []string) []string {
	var newUris []string
	portMap := make(map[string]int)
	// Add http/https ports if different from standard ports
	if dispatchConfig.APIHTTPPort != 80 {
		portMap["http"] = dispatchConfig.APIHTTPPort
	}
	if dispatchConfig.APIHTTPSPort != 443 {
		portMap["https"] = dispatchConfig.APIHTTPSPort
	}
	for _, uri := range uris {
		for _, protocol := range protocols {
			if port, ok := portMap[protocol]; ok {
				newUris = append(newUris, fmt.Sprintf("%s://%s:%d%s", protocol, dispatchConfig.Host, port, uri))
			} else {
				newUris = append(newUris, fmt.Sprintf("%s://%s%s", protocol, dispatchConfig.Host, uri))
			}

		}
	}
	return newUris
}
