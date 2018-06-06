///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	apiclient "github.com/vmware/dispatch/pkg/api-manager/gen/client/endpoint"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
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
			err := deleteAPI(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteAPI makes the API call to delete an API endpoint
func CallDeleteAPI(i interface{}) error {
	client := apiManagerClient()
	apiModel := i.(*v1.API)
	params := &apiclient.DeleteAPIParams{
		Context: context.Background(),
		API:     *apiModel.Name,
		Tags:    []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	deleted, err := client.Endpoint.DeleteAPI(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*apiModel = *deleted.Payload
	return nil
}

func deleteAPI(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	apiModel := v1.API{
		Name: &args[0],
	}
	err := CallDeleteAPI(&apiModel)
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
