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

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteFunctionLong = i18n.T(`Delete functions.`)

	// TODO: add examples
	deleteFunctionExample = i18n.T(``)
)

// NewCmdDeleteFunction creates command responsible for deleting functions.
func NewCmdDeleteFunction(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "function FUNCTION_NAME",
		Short:   i18n.T("Delete function"),
		Long:    deleteFunctionLong,
		Example: deleteFunctionExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"functions"},
		Run: func(cmd *cobra.Command, args []string) {
			c := functionManagerClient()
			err := deleteFunction(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteFunction makes the API call to delete a function
func CallDeleteFunction(c client.FunctionsClient) ModelAction {
	return func(i interface{}) error {
		functionModel := i.(*v1.Function)

		deleted, err := c.DeleteFunction(context.Background(), "", *functionModel.Name)
		if err != nil {
			return err
		}
		*functionModel = *deleted
		return nil
	}
}

func deleteFunction(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.FunctionsClient) error {
	functionModel := v1.Function{
		Name: &args[0],
	}
	err := CallDeleteFunction(c)(&functionModel)
	if err != nil {
		return err
	}
	return formatDeleteFunctionOutput(out, false, []*v1.Function{&functionModel})
}

func formatDeleteFunctionOutput(out io.Writer, list bool, functions []*v1.Function) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(functions)
		}
		return encoder.Encode(functions[0])
	}
	for _, f := range functions {
		_, err := fmt.Fprintf(out, "Deleted function: %s\n", *f.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
