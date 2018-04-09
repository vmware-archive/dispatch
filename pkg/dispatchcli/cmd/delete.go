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
	deleteLong = `Delete one or many resources.` + validResources

	deleteExample = i18n.T(`
		# Delete a single image with name "demo-python3-runtime"
		vs delete image demo-python3-runtime
		# Delete a single function with name "open-sesame"
		vs delete function open-sesame`)
)

// NewCmdDelete creates a command object for the generic "delete" action, which
// deletes one or more resources from a server.
func NewCmdDelete(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete TYPE [NAME|ID] [flags]",
		Short:   i18n.T("Delete one or many resources"),
		Long:    deleteLong,
		Example: deleteExample,
		Run: func(cmd *cobra.Command, args []string) {
			if file == "" {
				runHelp(cmd, args)
				return
			}

			deleteMap := map[string]modelAction{
				utils.ImageKind:     CallDeleteImage,
				utils.BaseImageKind: CallDeleteBaseImage,
				utils.FunctionKind:  CallDeleteFunction,
				utils.SecretKind:    CallDeleteSecret,
			}

			err := importFile(out, errOut, cmd, args, deleteMap)
			CheckErr(err)
		},
		SuggestFor: []string{"list"},
	}
	cmd.AddCommand(NewCmdDeleteBaseImage(out, errOut))
	cmd.AddCommand(NewCmdDeleteImage(out, errOut))
	cmd.AddCommand(NewCmdDeleteFunction(out, errOut))
	cmd.AddCommand(NewCmdDeleteSecret(out, errOut))
	cmd.AddCommand(NewCmdDeleteAPI(out, errOut))
	cmd.AddCommand(NewCmdDeleteSubscription(out, errOut))
	cmd.AddCommand(NewCmdDeleteEventDriver(out, errOut))
	cmd.AddCommand(NewCmdDeleteEventDriverType(out, errOut))
	cmd.AddCommand(NewCmdDeleteApplication(out, errOut))

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file")
	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")
	return cmd
}
