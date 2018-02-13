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
	updateLong = i18n.T(`Create a resource. See subcommands for resources that can be created.`)

	// TODO: Add examples
	updateExample = i18n.T(``)
)

// NewCmdUpdate creates command responsible for secret updates.
func NewCmdUpdate(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   i18n.T("Update resources."),
		Long:    updateLong,
		Example: updateExample,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" {
				runHelp(cmd, args)
				return
			}

			updateMap := map[string]modelAction{
				"secret": CallUpdateSecret,
			}

			err := importFile(out, errOut, cmd, args, updateMap)
			CheckErr(err)
		},
	}

	cmd.AddCommand(NewCmdUpdateSecret(out, errOut))
	cmd.AddCommand(NewCmdUpdateApplication(out, errOut))
	return cmd
}
