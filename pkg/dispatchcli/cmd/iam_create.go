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
	iamCreateLong = i18n.T(`Create iam resources. See subcommands for iam resources type`)

	// TODO: add examples
	iamCreateExample = i18n.T(``)
)

// NewCmdIamCreate creates a command object for the iam reources creation.
func NewCmdIamCreate(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   i18n.T("Create iam resources."),
		Long:    iamCreateLong,
		Example: iamCreateExample,
		Run:     runHelp,
	}
	cmd.AddCommand(NewCmdIamCreatePolicy(out, errOut))
	cmd.AddCommand(NewCmdIamCreateServiceAccount(out, errOut))
	cmd.AddCommand(NewCmdIamCreateOrganization(out, errOut))
	return cmd
}
