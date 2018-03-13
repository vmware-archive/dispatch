///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/utils"
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
				utils.APIKind:         CallUpdateAPI,
				utils.ApplicationKind: CallUpdateApplication,
				utils.BaseImageKind:   CallUpdateBaseImage,
				utils.SecretKind:      CallUpdateSecret,
				utils.PolicyKind:      CallUpdatePolicy,
			}

			err := importFile(out, errOut, cmd, args, updateMap)
			CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file")

	cmd.AddCommand(NewCmdUpdateSecret(out, errOut))
	cmd.AddCommand(NewCmdUpdateAPI(out, errOut))
	cmd.AddCommand(NewCmdUpdateApplication(out, errOut))
	cmd.AddCommand(NewCmdUpdateBaseImage(out, errOut))
	cmd.AddCommand(NewCmdUpdatePolicy(out, errOut))
	return cmd
}
