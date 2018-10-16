///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

// NewCmdCompletion creates a completion command for CLI
func NewCmdCompletion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: i18n.T("Generates bash completion scripts."),
		Run: func(cmd *cobra.Command, args []string) {
			cmds.GenBashCompletion(os.Stdout)
		},
	}
	return cmd
}
