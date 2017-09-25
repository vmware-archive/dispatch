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
	getLong = `Display one or many resources.` + validResources

	getExample = i18n.T(`
		# List all serverless functions.
		vs get functions
		# List a single image with name "demo-python3-runtime"
		vs get image demo-python3-runtime
		# List a single function with name "open-sesame"
		vs get function open-sesame`)
)

// NewCmdGet creates a command object for the generic "get" action, which
// retrieves one or more resources from a server.
func NewCmdGet(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get TYPE [NAME|ID] [flags]",
		Short:   i18n.T("Display one or many resources"),
		Long:    getLong,
		Example: getExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := runGet(out, errOut, cmd, args)
			CheckErr(err)
		},
		SuggestFor: []string{"list"},
	}
	cmd.AddCommand(NewCmdGetBaseImage(out, errOut))
	return cmd
}

func runGet(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}
