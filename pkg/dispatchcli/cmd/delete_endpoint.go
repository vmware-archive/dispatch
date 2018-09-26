///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteEndpointLong = i18n.T(`Delete api.`)

	// TODO: add examples
	deleteEndpointExample = i18n.T(``)
)

// NewCmdDeleteEndpoint creates command responsible for deleting Endpoint.
func NewCmdDeleteEndpoint(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "endpoint ENDPOINT_NAME",
		Short:   i18n.T("Delete endpoint"),
		Long:    deleteEndpointLong,
		Example: deleteEndpointExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"apis"},
		Run: func(cmd *cobra.Command, args []string) {
			c := endpointsClient()
			err := deleteEndpoint(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeleteEndpoint makes the Endpoint call to delete an Endpoint endpoint
func CallDeleteEndpoint(c client.EndpointsClient) ModelAction {
	return func(i interface{}) error {
		model := i.(*v1.Endpoint)

		deleted, err := c.DeleteEndpoint(context.TODO(), "", model.Name)
		if err != nil {
			return err
		}
		*model = *deleted
		return nil
	}
}

func deleteEndpoint(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EndpointsClient) error {
	apiModel := v1.Endpoint{
		Meta: v1.Meta{
			Name: args[0],
		},
	}
	err := CallDeleteEndpoint(c)(&apiModel)
	if err != nil {
		return err
	}
	return formatDeleteEndpointOutput(out, false, []*v1.Endpoint{&apiModel})
}

func formatDeleteEndpointOutput(out io.Writer, list bool, models []*v1.Endpoint) error {
	if w, err := formatOutput(out, list, models); w {
		return err
	}
	for _, m := range models {
		_, err := fmt.Fprintf(out, "Deleted api: %s\n", m.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
