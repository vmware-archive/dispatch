///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var (
	execLong = i18n.T(`Execute a serverless function.`)

	// TODO: Add examples
	execExample = i18n.T(``)
)

// NewCmdExec creates a command to execute a serverless function.
func NewCmdExec(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec [--secret SECRET_NAME] [--wait] [--params PARAMS] FUNCTION_NAME",
		Short:   i18n.T("Execute a serverless function"),
		Long:    execLong,
		Example: execExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := runExec(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func runExec(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}
