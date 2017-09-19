///////////////////////////////////////////////////////////////////////
// Copyright (C) 2016 VMware, Inc. All rights reserved.
// -- VMware Confidential
///////////////////////////////////////////////////////////////////////
package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"gitlab.eng.vmware.com/serverless/serverless/pkg/vscli/i18n"
)

var validResources = i18n.T(`Valid resource types include:
	* functions
	* images
	* base-images
	* secrets
    `)

// NewVSCLI creates cobra object for top-level VMware serverless CLI
func NewVSCLI(in io.Reader, out, errOut io.Writer) *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "vs",
		Short: i18n.T("vs allows to interact with VMware Serverless platform."),
		Long:  i18n.T("vs allows to interact with VMware Serverless platform."),
		Run:   runHelp,
	}

	cmds.AddCommand(NewCmdGet(out, errOut))
	cmds.AddCommand(NewCmdCreate(out, errOut))
	cmds.AddCommand(NewCmdExec(out, errOut))
	cmds.AddCommand(NewCmdLogin(in, out, errOut))
	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}
