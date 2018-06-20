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
	iamGetLong = i18n.T(`Get iam resources. See subcommands for iam resources type`)

	// TODO: add examples
	iamGetExample = i18n.T(``)
)

// NewCmdIamGet creates a command object for the iam reources creation.
func NewCmdIamGet(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   i18n.T("Get iam resources."),
		Long:    iamGetLong,
		Example: iamGetExample,
		Run:     runHelp,
	}
	cmd.AddCommand(NewCmdIamGetPolicy(out, errOut))
	cmd.AddCommand(NewCmdIamGetServiceAccount(out, errOut))
	cmd.AddCommand(NewCmdIamGetOrganization(out, errOut))
	return cmd
}
