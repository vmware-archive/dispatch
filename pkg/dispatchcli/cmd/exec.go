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

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/client/runner"
	models "github.com/vmware/dispatch/pkg/function-manager/gen/models"
)

var (
	execLong = i18n.T(`Execute a serverless function.`)

	// TODO: Add examples
	execExample = i18n.T(``)

	execWait    = false
	execInput   = "{}"
	execSecrets = []string{}
)

// NewCmdExec creates a command to execute a serverless function.
func NewCmdExec(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec [--wait] [--input JSON] [--secret SECRET_1,SECRET_2...] FUNCTION_NAME",
		Short:   i18n.T("Execute a serverless function"),
		Long:    execLong,
		Example: execExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := runExec(out, errOut, cmd, args)
			CheckErr(err)
		},
		PreRunE: validateFnExecFunc(errOut),
	}
	cmd.Flags().BoolVar(&execWait, "wait", false, "Wait for the function to complete execution.")
	cmd.Flags().StringVar(&execInput, "input", "{}", "Function input JSON object")
	cmd.Flags().StringArrayVar(&execSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")
	return cmd
}

func validateFnExecFunc(errOut io.Writer) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func runExec(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	functionName := args[0]
	var input map[string]interface{}
	err := json.Unmarshal([]byte(execInput), &input)
	if err != nil {
		fmt.Fprintf(errOut, "Error when parsing function parameters %s\n", execInput)
		return err
	}
	run := &models.Run{
		Blocking: execWait,
		Input:    input,
		Secrets:  execSecrets,
	}

	params := &fnrunner.RunFunctionParams{
		Body:         run,
		Context:      context.Background(),
		FunctionName: functionName,
	}
	client := functionManagerClient()
	executed, executing, err := client.Runner.RunFunction(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if executed != nil {
		if !dispatchConfig.Json {
			fmt.Fprintf(out, "Function %s finished successfully.\n", functionName)
		}
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(executed.Payload.Output)
	}
	if executing != nil {
		// TODO (imikushin): need to return a run ID and support JSON
		fmt.Fprintf(out, "Function %s started\n", functionName)
	}
	return nil
}
