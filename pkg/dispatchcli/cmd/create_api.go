///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/client"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
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

	httpsOnly = false
	disable   = false
	cors      = false
	hosts     = []string{}
	paths     = []string{"/"}
	methods   = []string{"GET"}
	auth      = "public"
)

// NewCmdCreateAPI creates command responsible for dispatch function api creation.
func NewCmdCreateAPI(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api API_NAME FUNCTION_NAME [--auth AUTH_METHOD] [--domain DOMAINNAME...] [--method METHOD...] [--path PATH...] [--disable] [--cors] [--https-only]",
		Short:   i18n.T("Create api"),
		Long:    createAPILong,
		Example: createAPIExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			c := apiManagerClient()
			err := createAPI(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringSliceVarP(&hosts, "domain", "d", []string{}, "domain names that point to your API. Use multiple times or separate with commas to specify more values.")
	cmd.Flags().StringSliceVarP(&paths, "path", "p", []string{"/"}, "relative paths that point to your API. Use multiple times or separate with commas to specify more values.")
	cmd.Flags().StringSliceVarP(&methods, "method", "m", []string{"GET"}, "methods that point to your API. Use multiple times or separate with commas to specify more values.")
	cmd.Flags().BoolVar(&httpsOnly, "https-only", false, "only support https connections")
	cmd.Flags().BoolVar(&disable, "disable", false, "disable the api")
	cmd.Flags().BoolVar(&cors, "cors", false, "enable CORS")
	cmd.Flags().StringVar(&auth, "auth", "public", "specify end-user authentication method, (e.g. public, basic, oauth2)")
	return cmd
}

// CallCreateAPI makes the API call to create an API endpoint
func CallCreateAPI(c client.APIsClient) ModelAction {
	return func(f interface{}) error {
		api := f.(*v1.API)

		created, err := c.CreateAPI(context.TODO(), "", api)
		if err != nil {
			return err
		}
		*api = *created
		return nil
	}
}

func createAPI(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.APIsClient) error {

	apiName := args[0]
	function := args[1]

	protocols := []string{"http", "https"}
	if httpsOnly {
		protocols = []string{"https"}
	}

	api := &v1.API{
		Name:           swag.String(apiName),
		Function:       swag.String(function),
		Protocols:      protocols,
		Methods:        methods,
		Uris:           paths,
		Hosts:          hosts,
		Authentication: auth,
		Enabled:        !disable,
		Cors:           cors,
		Tags:           []*v1.Tag{},
	}
	if cmdFlagApplication != "" {
		api.Tags = append(api.Tags, &v1.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	err := CallCreateAPI(c)(api)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, api); w {
		return err
	}
	fmt.Fprintf(out, "Created api: %s\n", *api.Name)
	return nil
}
