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

	"github.com/pkg/errors"
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

	filePath     string
	httpsOnlyArr []string
	disableArr   []string
	corsArr      []string
	authArr      []string
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
		Use:     "api API_NAME",
		Short:   i18n.T("Update api"),
		Long:    updateAPILong,
		Example: updateAPIExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updateAPI(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "path to file to use as update")
	cmd.Flags().StringArrayVarP(&hosts, "domain", "d", []string{"!"}, "domain names that point to your API (multi-values)")
	cmd.Flags().StringArrayVarP(&paths, "path", "p", []string{}, "paths that point to your API (multi-values)")
	cmd.Flags().StringArrayVarP(&methods, "method", "m", []string{}, "methods that point to your API")
	cmd.Flags().StringArrayVar(&httpsOnlyArr, "https-only", []string{}, "only support https connections")
	cmd.Flags().StringArrayVar(&disableArr, "disable", []string{}, "disable the api")
	cmd.Flags().StringArrayVar(&corsArr, "cors", []string{}, "enable CORS")
	cmd.Flags().StringArrayVarP(&authArr, "auth", "a", []string{}, "specify end-user authentication method, (e.g. public, basic, oauth2)")
	return cmd
}

func updateAPI(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {
	apiName := args[0]
	api := models.API{}

	if filePath != "" && (hosts[0] != "!" || len(paths) != 0 || len(methods) != 0 || len(httpsOnlyArr) != 0 || len(disableArr) != 0 || len(corsArr) != 0 || len(authArr) != 0) {
		apiContent, err := ioutil.ReadFile(filePath)
		if err != nil {
			message := fmt.Sprintf("Error when reading content of %s", filePath)
			return formatCliError(err, message)
		}
		if err := json.Unmarshal(apiContent, &api); err != nil {
			message := fmt.Sprintf("Error when parsing JSON from %s with error %s", apiContent, err)
			return formatCliError(err, message)
		}

		if apiName != *api.Name {
			return formatCliError(errors.New("name mismatch"), "command line arg API_NAME does not match name in file")
		}

		err = CallUpdateAPI(&api)
		if err != nil {
			return err
		}

	} else if filePath == "" && (hosts[0] != "!" || len(paths) != 0 || len(methods) != 0 || len(httpsOnlyArr) != 0 || len(disableArr) != 0 || len(corsArr) != 0 || len(authArr) != 0) {
		//updatePartialAPI(apiName)
	}

	fmt.Fprintf(out, "Updated api: %s\n", *api.Name)
	return nil
}

func updatePartialAPI(apiName string) error {
	api := models.API{}

	err := CallUpdateAPI(&api)
	if err != nil {
		return err
	}

	return nil
}
