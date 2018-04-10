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
	iamDeleteLong = i18n.T(`Delete iam resources. See subcommands for iam resources type`)

	// TODO: add examples
	iamDeleteExample = i18n.T(``)
)

// NewCmdIamDelete creates a command object for the iam reources creation.
func NewCmdIamDelete(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   i18n.T("Delete iam resources."),
		Long:    iamDeleteLong,
		Example: iamDeleteExample,
		Run:     runHelp,
	}
	cmd.AddCommand(NewCmdIamDeletePolicy(out, errOut))
	cmd.AddCommand(NewCmdIamDeleteServiceAccount(out, errOut))
	return cmd
}
