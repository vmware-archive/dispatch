///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	getLong = `Display one or many resources.` + validResources

	getExample = i18n.T(`
		# List all dispatch functions.
		dispatch get functions
		# List a single image with name "demo-python3-runtime"
		dispatch get image demo-python3-runtime
		# List a single function with name "open-sesame"
		dispatch get function open-sesame`)
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
	cmd.AddCommand(NewCmdGetImage(out, errOut))
	cmd.AddCommand(NewCmdGetFunction(out, errOut))
	cmd.AddCommand(NewCmdGetRun(out, errOut))
	cmd.AddCommand(NewCmdGetSecret(out, errOut))
	cmd.AddCommand(NewCmdGetEndpoint(out, errOut))
	cmd.AddCommand(NewCmdGetSubscription(out, errOut))
	cmd.AddCommand(NewCmdGetEventDriver(out, errOut))
	cmd.AddCommand(NewCmdGetEventDriverType(out, errOut))
	cmd.AddCommand(NewCmdGetApplication(out, errOut))
	cmd.AddCommand(NewCmdGetServiceClass(out, errOut))
	cmd.AddCommand(NewCmdGetServiceInstance(out, errOut))
	return cmd
}

func runGet(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	runHelp(cmd, args)
	return nil
}
