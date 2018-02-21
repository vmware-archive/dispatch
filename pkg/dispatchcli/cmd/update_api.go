///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/api-manager/gen/client/endpoint"
	"github.com/vmware/dispatch/pkg/api-manager/gen/models"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	updateAPILong    = "update an api based on json"
	updateAPIExample = `dispatch update api my_api --path /new/path`

	httpsOnlyStr string
	disableStr   string
	corsStr      string
)

// CallUpdateAPI makes the backend service call to update an api
func CallUpdateAPI(input interface{}) error {
	apiBody := input.(*models.API)

	params := endpoint.NewUpdateAPIParams()
	params.API = *apiBody.Name
	params.Body = apiBody

	_, err := apiManagerClient().Endpoint.UpdateAPI(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

// NewCmdUpdateAPI creates command responsible for updating an api
func NewCmdUpdateAPI(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api API_NAME [--function FUNCTION_NAME] [--auth AUTH_METHOD] [--domain DOMAINNAME...] [--method METHOD...] [--path PATH...] [--disable] [--cors] [--https-only]",
		Short:   i18n.T("Update api"),
		Long:    updateAPILong,
		Example: updateAPIExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateAPI(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVar(&functionName, "function", "", "associate api with a function")
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringArrayVarP(&hosts, "domain", "d", []string{}, "domain names that point to your API (multi-values)")
	cmd.Flags().StringArrayVarP(&paths, "path", "p", []string{"/"}, "paths that point to your API (multi-values)")
	cmd.Flags().StringArrayVarP(&methods, "method", "m", []string{"GET"}, "methods that point to your API")
	cmd.Flags().StringVar(&httpsOnlyStr, "https-only", "false", "only support https connections")
	cmd.Flags().StringVar(&disableStr, "disable", "false", "disable the api")
	cmd.Flags().StringVar(&corsStr, "cors", "false", "enable CORS")
	cmd.Flags().StringVar(&auth, "auth", "public", "specify end-user authentication method, (e.g. public, basic, oauth2)")
	return cmd
}

func updateAPI(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	apiName := args[0]

	params := endpoint.NewGetAPIParams()
	params.API = apiName
	apiOk, err := apiManagerClient().Endpoint.GetAPI(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	api := *apiOk.Payload
	changed := false

	if cmd.Flags().Changed("function") {
		api.Function = &functionName
		changed = true
	}

	if cmd.Flags().Changed("application") {
		api.Tags = append(models.APITags{}, &models.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
		changed = true
	}

	if cmd.Flags().Changed("domain") {
		api.Hosts = hosts
		changed = true
	}

	if cmd.Flags().Changed("path") {
		api.Uris = paths
		changed = true
	}

	if cmd.Flags().Changed("method") {
		api.Methods = methods
		changed = true
	}

	if cmd.Flags().Changed("https-only") {
		httpsOnly, err := strconv.ParseBool(httpsOnlyStr)
		if err != nil {
			return formatCliError(err, fmt.Sprintf("Failed parsing https-only value: %s", httpsOnlyStr))
		}

		protocols := []string{"https"}
		if !httpsOnly {
			protocols = append(protocols, "http")
		}
		api.Protocols = protocols
		changed = true
	}

	if cmd.Flags().Changed("disable") {
		disable, err := strconv.ParseBool(disableStr)
		if err != nil {
			return formatCliError(err, fmt.Sprintf("Failed parsing disable value: %s", disableStr))
		}

		api.Enabled = !disable
		changed = true
	}

	if cmd.Flags().Changed("cors") {
		cors, err := strconv.ParseBool(corsStr)
		if err != nil {
			return formatCliError(err, fmt.Sprintf("Failed parsint cors value: %s", corsStr))
		}

		api.Cors = cors
		changed = true
	}

	if cmd.Flags().Changed("auth") {
		api.Authentication = auth
		changed = true
	}

	if !changed {
		fmt.Fprintf(out, "No fields changed\n")
		return nil
	}

	err = CallUpdateAPI(&api)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated api: %s\n", *api.Name)
	return nil
}
