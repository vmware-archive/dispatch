///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/swag"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	apiclient "github.com/vmware/dispatch/pkg/api-manager/gen/client/endpoint"
	"github.com/vmware/dispatch/pkg/api-manager/gen/models"
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
			if len(args) == 1 {
				err = getAPI(out, errOut, cmd, args)
			} else {
				err = getAPIs(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&functionName, "func", "f", "", "get all apis for specified function")
	return cmd
}

func getAPIs(out, errOut io.Writer, cmd *cobra.Command) error {

	params := &apiclient.GetApisParams{
		Context: context.Background(),
	}
	if functionName != "" {
		params.Function = swag.String(functionName)
	}

	client := apiManagerClient()
	get, err := client.Endpoint.GetApis(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatAPIOutput(out, true, get.Payload)
}

func getAPI(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	apiName := args[0]
	params := &apiclient.GetAPIParams{
		API:     apiName,
		Context: context.Background(),
	}
	client := apiManagerClient()

	get, err := client.Endpoint.GetAPI(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return formatAPIOutput(out, false, []*models.API{get.Payload})
}

func formatAPIOutput(out io.Writer, list bool, apis []*models.API) error {

	if dispatchConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(apis)
		}
		return encoder.Encode(apis[0])
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
			strings.Join(a.Uris, "\n"), a.Authentication, string(a.Status), fmt.Sprintf("%t", a.Enabled),
		})
	}
	table.Render()
	return nil
}
