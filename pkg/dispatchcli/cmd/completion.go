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

var (
	completionLong    = i18n.T(`Generates bash completion scripts.`)
	completionExample = i18n.T(`
		# Installing bash completion on Linux
		## Load the dispatch completion code for bash into the current shell.
		source <(dispatch completion)

		# Installing bash completion on MacOS with Homebrew
		## Process substitution doesn't work in the bash bundled with MacOS.
		## You have to save it to a file and read it in.
		brew install bash-completion
		## Edit .bash_profile as instructed.
		dispatch completion > $(brew --prefix)/etc/bash_completion.d/dispatch
	`)
)

// NewCmdCompletion creates a completion command for CLI
func NewCmdCompletion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "completion",
		Short:   i18n.T("Generates bash completion scripts."),
		Long:    completionLong,
		Example: completionExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmds.GenBashCompletion(os.Stdout)
		},
	}
	return cmd
}
