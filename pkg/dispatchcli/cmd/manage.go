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
	manageShort = `Manage Dispatch configurations.`
	manageLong  = `Manage Dispatch configurations.`
)

// NewCmdManage creates a command object for Dispatch "manage" action
func NewCmdManage(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   i18n.T(`manage`),
		Short: manageShort,
		Long:  manageLong,
		Run:   runHelp,
	}

	cmd.AddCommand(NewCmdManageBootstrap(out, errOut))
	return cmd
}
