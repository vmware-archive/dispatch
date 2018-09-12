///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
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

	contentType = ""
	accept      = ""
	execSecrets = []string{}
)

// NewCmdExec creates a command to execute a dispatch function.
func NewCmdExec(in io.Reader, out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec FUNCTION_NAME [flags] < in.json > out.json",
		Short:   i18n.T("Execute a dispatch function"),
		Long:    execLong,
		Example: execExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := functionManagerClient()
			err := runExec(in, out, errOut, cmd, args, c)
			CheckErr(err)
		},
		PreRunE: validateFnExecFunc(errOut),
	}
	cmd.Flags().StringArrayVar(&execSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")
	cmd.Flags().StringVarP(&contentType, "content-type", "c", "application/json", "Input Content-Type")
	cmd.Flags().StringVarP(&accept, "accept", "a", "application/json", "Output Content-Type")
	return cmd
}

func validateFnExecFunc(errOut io.Writer) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

func runExec(in io.Reader, out, errOut io.Writer, cmd *cobra.Command, args []string, c client.FunctionsClient) error {
	functionName := args[0]

	inBytes, err := ioutil.ReadAll(in)
	if err != nil {
		return errors.Wrap(err, "reading stdin")
	}
	run := &v1.Run{
		Blocking:     true,
		InputBytes:   inBytes,
		HTTPContext:  map[string]interface{}{"Content-Type": contentType, "Accept": accept},
		Secrets:      execSecrets,
		FunctionName: functionName,
	}

	functionResult, err := c.RunFunction(context.TODO(), "", run)

	if err != nil {
		return errors.Wrap(err, "api client error")
	}

	out.Write(functionResult.OutputBytes)

	return nil
}

func formatExecOutput(out io.Writer, run *v1.Run) error {
	// Always return json for execution
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "    ")
	return encoder.Encode(run)
}
