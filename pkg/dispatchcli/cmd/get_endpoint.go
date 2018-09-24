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
	getEndpointLong = i18n.T(
		`Get dispatch function endpoint.`)
	// TODO: add examples
	getEndpointExample = i18n.T(``)

	functionName = ""
)

// NewCmdGetEndpoint gets command responsible for dispatch function endpoint creation.
func NewCmdGetEndpoint(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "endpoint [ENDPOINT_NAME] [--func FUNC_NAME]",
		Short:   i18n.T("Get endpoint"),
		Long:    getEndpointLong,
		Example: getEndpointExample,
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := endpointsClient()
			if len(args) == 1 {
				err = getEndpoint(out, errOut, cmd, args, c)
			} else {
				err = getEndpoints(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&functionName, "func", "f", "", "get all apis for specified function")
	return cmd
}

func getEndpoints(out, errOut io.Writer, cmd *cobra.Command, c client.EndpointsClient) error {
	get, err := c.ListEndpoints(context.TODO(), "")
	if err != nil {
		return err
	}
	return formatEndpointOutput(out, true, get)
}

func getEndpoint(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EndpointsClient) error {

	apiName := args[0]
	get, err := c.GetEndpoint(context.TODO(), "", apiName)
	if err != nil {
		return err
	}

	return formatEndpointOutput(out, false, []v1.Endpoint{*get})
}

func formatEndpointOutput(out io.Writer, list bool, apis []v1.Endpoint) error {

	if w, err := formatOutput(out, list, apis); w {
		return err
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Function", "Protocol", "Method", "Domain", "Path", "Enabled"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("-")
	table.SetRowLine(true)
	for _, a := range apis {
		table.Append([]string{
			a.Name, a.Function,
			strings.Join(a.Protocols, "\n"), strings.Join(a.Methods, "\n"), strings.Join(a.Hosts, "\n"),
			strings.Join(a.Uris, "\n"), fmt.Sprintf("%t", a.Enabled),
		})
	}
	table.Render()
	return nil
}
