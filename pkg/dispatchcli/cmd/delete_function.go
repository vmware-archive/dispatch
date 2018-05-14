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
	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	function "github.com/vmware/dispatch/pkg/function-manager/gen/client/store"
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
			err := deleteFunction(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteFunction makes the API call to delete a function
func CallDeleteFunction(i interface{}) error {
	client := functionManagerClient()
	functionModel := i.(*v1.Function)
	params := &function.DeleteFunctionParams{
		FunctionName: *functionModel.Name,
		Context:      context.Background(),
		Tags:         []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	deleted, err := client.Store.DeleteFunction(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*functionModel = *deleted.Payload
	return nil
}

func deleteFunction(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	functionModel := v1.Function{
		Name: &args[0],
	}
	err := CallDeleteFunction(&functionModel)
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
