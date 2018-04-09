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
	iamDescription = i18n.T(`Managing identity and access management (iam) resources.`)

	// TODO: Add examples
	iamExample = i18n.T(``)
)

// NewCmdIam creates a command for iam resource management
func NewCmdIam(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: iamDescription,
		Long:  iamDescription,
		Run:   runHelp,
	}
	cmd.AddCommand(NewCmdIamCreate(out, errOut))
	cmd.AddCommand(NewCmdIamGet(out, errOut))
	cmd.AddCommand(NewCmdIamDelete(out, errOut))
	return cmd
}
