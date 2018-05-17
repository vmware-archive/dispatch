///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	execLong = i18n.T(`Execute a dispatch function.`)

	// TODO: Add examples
	execExample = i18n.T(``)

	execWait      = false
	execAllOutput = false
	execInput     = "{}"
	execSecrets   = []string{}
)

// NewCmdExec creates a command to execute a dispatch function.
func NewCmdExec(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec [--wait] [--input JSON] [--secret SECRET_1,SECRET_2...] FUNCTION_NAME",
		Short:   i18n.T("Execute a dispatch function"),
		Long:    execLong,
		Example: execExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := functionManagerClient()
			err := runExec(out, errOut, cmd, args, c)
			CheckErr(err)
		},
		PreRunE: validateFnExecFunc(errOut),
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	cmd.Flags().BoolVar(&execWait, "wait", false, "Wait for the function to complete execution.")
	cmd.Flags().StringVar(&execInput, "input", "{}", "Function input JSON object")
	cmd.Flags().StringArrayVar(&execSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")
	cmd.Flags().BoolVar(&execAllOutput, "all", false, "Also print metadata along with json output, ONLY with --json")
	return cmd
}

func validateFnExecFunc(errOut io.Writer) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func runExec(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.FunctionsClient) error {
	functionName := args[0]
	var input interface{}
	err := json.Unmarshal([]byte(execInput), &input)
	if err != nil {
		fmt.Fprintf(errOut, "Error when parsing function parameters %s\n", execInput)
		return err
	}
	run := &v1.Run{
		Blocking:     execWait,
		Input:        input,
		Secrets:      execSecrets,
		FunctionName: functionName,
	}

	functionResult, err := c.RunFunction(context.TODO(), "", run)

	if err != nil {
		return formatAPIError(err, run)
	}

	return formatExecOutput(out, functionResult)
}

func formatExecOutput(out io.Writer, run *v1.Run) error {
	// Always return json for execution
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "    ")
	return encoder.Encode(run)
}
