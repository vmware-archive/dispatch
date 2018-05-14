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
	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	fnrunner "github.com/vmware/dispatch/pkg/function-manager/gen/client/runner"
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
			err := runExec(out, errOut, cmd, args)
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

func runExec(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	functionName := args[0]
	var input interface{}
	err := json.Unmarshal([]byte(execInput), &input)
	if err != nil {
		fmt.Fprintf(errOut, "Error when parsing function parameters %s\n", execInput)
		return err
	}
	run := &v1.Run{
		Blocking: execWait,
		Input:    input,
		Secrets:  execSecrets,
	}

	params := &fnrunner.RunFunctionParams{
		Body:         run,
		Context:      context.Background(),
		FunctionName: &functionName,
		Tags:         []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	client := functionManagerClient()
	executed, executing, err := client.Runner.RunFunction(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if executed != nil {
		return formatExecOutput(out, executed.Payload)
	} else if executing != nil {
		return formatExecOutput(out, executing.Payload)
	} else {
		// We should never get here... just in case
		return fmt.Errorf("Unexepected response from API")
	}
}

func formatExecOutput(out io.Writer, run *v1.Run) error {
	// Always return json for execution
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "    ")
	return encoder.Encode(run)
}
