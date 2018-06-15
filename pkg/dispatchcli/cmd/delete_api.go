///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteAPILong = i18n.T(`Delete api.`)

	// TODO: add examples
	deleteAPIExample = i18n.T(``)
)

// NewCmdDeleteAPI creates command responsible for deleting API.
func NewCmdDeleteAPI(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api API_NAME",
		Short:   i18n.T("Delete API"),
		Long:    deleteAPILong,
		Example: deleteAPIExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"apis"},
		Run: func(cmd *cobra.Command, args []string) {
			c := apiManagerClient()
			err := deleteAPI(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteAPI makes the API call to delete an API endpoint
func CallDeleteAPI(c client.APIsClient) ModelAction {
	return func(i interface{}) error {
		apiModel := i.(*v1.API)

		deleted, err := c.DeleteAPI(context.TODO(), "", *apiModel.Name)
		if err != nil {
			return err
		}
		*apiModel = *deleted
		return nil
	}
}

func deleteAPI(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.APIsClient) error {
	apiModel := v1.API{
		Name: &args[0],
	}
	err := CallDeleteAPI(c)(&apiModel)
	if err != nil {
		return err
	}
	return formatDeleteAPIOutput(out, false, []*v1.API{&apiModel})
}

func formatDeleteAPIOutput(out io.Writer, list bool, apis []*v1.API) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(apis)
		}
		return encoder.Encode(apis[0])
	}
	for _, a := range apis {
		_, err := fmt.Fprintf(out, "Deleted api: %s\n", *a.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
