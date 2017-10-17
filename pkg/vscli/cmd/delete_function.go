///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	function "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client/store"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
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
	return cmd
}

func deleteFunction(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := functionManagerClient()
	params := &function.DeleteFunctionParams{
		Context:      context.Background(),
		FunctionName: args[0],
	}
	resp, err := client.Store.DeleteFunction(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatDeleteFunctionOutput(out, false, []*models.Function{resp.Payload})
}

func formatDeleteFunctionOutput(out io.Writer, list bool, functions []*models.Function) error {
	if vsConfig.Json {
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
