///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	apiclient "github.com/vmware/dispatch/pkg/api-manager/gen/client/endpoint"
	models "github.com/vmware/dispatch/pkg/api-manager/gen/models"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createAPILong = i18n.T(
		`Create dispatch function api.

Note:
  Import your own tls certificates if you want to use your own domain name with HTTPS secure connection
		`)
	// TODO: add examples
	createAPIExample = i18n.T(``)

	httpsOnly            = false
	disable              = false
	cors                 = false
	hosts                = []string{}
	paths                = []string{"/"}
	methods              = []string{"GET"}
	auth                 = "public"
	createAPIApplication = i18n.T(``)
)

// NewCmdCreateAPI creates command responsible for dispatch function api creation.
func NewCmdCreateAPI(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api API_NAME FUNCTION_NAME [--auth AUTH_METHOD] [--domain DOMAINNAME...] [--method METHOD...] [--path PATH...] [--disable] [--cors] [--https] [--https-only]",
		Short:   i18n.T("Create api"),
		Long:    createAPILong,
		Example: createAPIExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := createAPI(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringArrayVarP(&hosts, "domain", "d", []string{}, "domain names that point to your API (multi-values), default: empty")
	cmd.Flags().StringArrayVarP(&paths, "path", "p", []string{"/"}, "paths that point to your API (multi-values), default: /")
	cmd.Flags().StringArrayVarP(&methods, "method", "m", []string{"GET"}, "methods that point to your API, default: GET")
	cmd.Flags().BoolVar(&httpsOnly, "https-only", false, "only support https connections, default: false")
	cmd.Flags().BoolVar(&disable, "disable", false, "disable the api, default: false")
	cmd.Flags().BoolVar(&cors, "cors", false, "enable CORS, default: false")
	cmd.Flags().StringVar(&auth, "auth", "public", "specify end-user authentication method, (e.g. public, basic, oauth2), default: public")
	return cmd
}

func createAPI(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	apiName := args[0]
	function := args[1]

	protocols := []string{"http", "https"}
	if httpsOnly {
		protocols = []string{"https"}
	}

	api := &models.API{
		Name:           swag.String(apiName),
		Function:       swag.String(function),
		Protocols:      protocols,
		Methods:        methods,
		Uris:           paths,
		Hosts:          hosts,
		Authentication: auth,
		Enabled:        !disable,
		Cors:           cors,
		Tags:           []*models.Tag{},
	}
	if cmdFlagApplication != "" {
		api.Tags = append(api.Tags, &models.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	params := &apiclient.AddAPIParams{
		Body:    api,
		Context: context.Background(),
	}
	client := apiManagerClient()

	created, err := client.Endpoint.AddAPI(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if dispatchConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Fprintf(out, "Created api: %s\n", *created.Payload.Name)
	return nil
}
