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
	deleteLong = `Delete one or many resources.` + validResources

	deleteExample = i18n.T(`
		# Delete a single image with name "demo-python3-runtime"
		vs delete image demo-python3-runtime
		# Delete a single function with name "open-sesame"
		vs delete function open-sesame`)
)

// NewCmdDelete creates a command object for the generic "delete" action, which
// deletes one or more resources from a server.
func NewCmdDelete(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete TYPE [NAME|ID] [flags]",
		Short:   i18n.T("Delete one or many resources"),
		Long:    deleteLong,
		Example: deleteExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := runDelete(out, errOut, cmd, args)
			CheckErr(err)
		},
		SuggestFor: []string{"list"},
	}
	cmd.AddCommand(NewCmdDeleteBaseImage(out, errOut))
	cmd.AddCommand(NewCmdDeleteImage(out, errOut))
	cmd.AddCommand(NewCmdDeleteFunction(out, errOut))
	cmd.AddCommand(NewCmdDeleteSecret(out, errOut))
	return cmd
}

func runDelete(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	return errors.New("Not implemented")
}
