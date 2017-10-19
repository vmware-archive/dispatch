///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	fnrunner "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/client/runner"
	models "gitlab.eng.vmware.com/serverless/serverless/pkg/function-manager/gen/models"
	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	execLong = i18n.T(`Execute a serverless function.`)

	// TODO: Add examples
	execExample = i18n.T(``)

	execWait    = false
	execInput   = "{}"
	execSecrets = "[]"
)

// NewCmdExec creates a command to execute a serverless function.
func NewCmdExec(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec [--wait] [--input JSON] [--secret JSON Array] FUNCTION_NAME",
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
	cmd.Flags().StringVar(&execSecrets, "secrets", "[]", "Function secret JSON array")
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
	var secrets []string
	err = json.Unmarshal([]byte(execSecrets), &secrets)
	if err != nil {
		fmt.Fprintf(errOut, "Error when parsing function secrets %s\n", execSecrets)
		return err
	}
	run := &models.Run{
		Blocking: execWait,
		Input:    input,
		Secrets:  secrets,
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
		if !vsConfig.Json {
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
