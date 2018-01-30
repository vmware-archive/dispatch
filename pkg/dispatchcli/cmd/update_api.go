///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/api-manager/gen/client/endpoint"
	"github.com/vmware/dispatch/pkg/api-manager/gen/models"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	updateAPILong    = "update an api based on json"
	updateAPIExample = `Example json
	{
		"authentication": "public",
		"enabled": true,
		"function": "hello-py",
		"hosts": [
			"example.com"
		],
		"methods": [
			"GET"
		],
		"name": "hello-api",
		"protocols": [
			"http",
			"https"
		],
		"tags": null,
		"uris": [
			"/hello"
		]
	}`
)

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

func NewCmdUpdateAPI(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api API_FILE",
		Short:   i18n.T("Update api"),
		Long:    updateAPILong,
		Example: updateAPIExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateAPI(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	return cmd
}

func updateAPI(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	filePath := args[0]
	api := models.API{}

	if filePath != "" {
		apiContent, err := ioutil.ReadFile(filePath)
		if err != nil {
			message := fmt.Sprintf("Error when reading content of %s", filePath)
			return formatCliError(err, message)
		}
		if err := json.Unmarshal(apiContent, &api); err != nil {
			message := fmt.Sprintf("Error when parsing JSON from %s with error %s", apiContent, err)
			return formatCliError(err, message)
		}
	}

	err := CallUpdateAPI(&api)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated api: %s\n", *api.Name)
	return nil
}
