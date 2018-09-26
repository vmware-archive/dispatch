///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/client"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createEndpointLong = i18n.T(
		`Create dispatch function endpoint.

Note:
  Import your own tls certificates if you want to use your own domain name with HTTPS secure connection
		`)
	// TODO: add examples
	createEndpointExample = i18n.T(``)

	httpsOnly = false
	disable   = false
	cors      = false
	hosts     = []string{}
	paths     = []string{"/"}
	methods   = []string{"GET"}
	auth      = "public"
)

// NewCmdCreateAPI creates command responsible for dispatch function endpoint creation.
func NewCmdCreateAPI(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "endpoint ENDPOINT_NAME FUNCTION_NAME [--auth AUTH_METHOD] [--domain DOMAINNAME...] [--method METHOD...] [--path PATH...] [--disable] [--cors] [--https-only]",
		Short:   i18n.T("Create endpoint"),
		Long:    createEndpointLong,
		Example: createEndpointExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			c := endpointsClient()
			err := createEndpoint(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}

	cmd.Flags().StringArrayVarP(&hosts, "domain", "d", []string{}, "domain names that point to your API (multi-values), default: empty")
	cmd.Flags().StringArrayVarP(&paths, "path", "p", []string{"/"}, "relative paths that point to your API (multi-values), default: /")
	cmd.Flags().StringArrayVarP(&methods, "method", "m", []string{"GET"}, "methods that point to your API, default: GET")
	cmd.Flags().BoolVar(&httpsOnly, "https-only", false, "only support https connections, default: false")
	cmd.Flags().BoolVar(&disable, "disable", false, "disable the api, default: false")
	cmd.Flags().BoolVar(&cors, "cors", false, "enable CORS, default: false")
	cmd.Flags().StringVar(&auth, "auth", "public", "specify end-user authentication method, (e.g. public, basic, oauth2), default: public")
	return cmd
}

// CallCreateEndpoint makes the API call to create an endpoint
func CallCreateEndpoint(c client.EndpointsClient) ModelAction {
	return func(f interface{}) error {
		model := f.(*v1.Endpoint)

		created, err := c.CreateEndpoint(context.TODO(), "", model)
		if err != nil {
			return err
		}
		*model = *created
		return nil
	}
}

func createEndpoint(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EndpointsClient) error {

	name := args[0]
	function := args[1]

	protocols := []string{"http", "https"}
	if httpsOnly {
		protocols = []string{"https"}
	}

	model := &v1.Endpoint{
		Meta: v1.Meta{
			Kind: v1.EndpointKind,
			Name: name,
			Tags: []*v1.Tag{},
		},
		Function:  function,
		Protocols: protocols,
		Methods:   methods,
		Uris:      paths,
		Hosts:     hosts,
		Enabled:   !disable,
		Cors:      cors,
	}

	err := CallCreateEndpoint(c)(model)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, model); w {
		return err
	}
	fmt.Fprintf(out, "Created endpoint: %s\n", model.Name)
	return nil
}
