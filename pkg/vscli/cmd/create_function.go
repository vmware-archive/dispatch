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
	createFunctionLong = i18n.T(`Create serverless function.`)

	// TODO: add examples
	createFunctionExample = i18n.T(``)
)

// NewCmdCreateFunction creates command responsible for serverless function creation.
func NewCmdCreateFunction(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "function [--name IMAGE_NAME] [--schema-in SCHEMA_FILE] [--image IMAGE_NAME] FILE",
		Short:   i18n.T("Create function"),
		Long:    createFunctionLong,
		Example: createFunctionExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := createFunction(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func createFunction(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}
