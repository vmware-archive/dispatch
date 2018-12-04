///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"
	"strings"

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

var deleteMap map[string]ModelAction

func initDeleteMap() {
	fnClient := functionManagerClient()
	imgClient := imageManagerClient()
	eventClient := eventManagerClient()
	apiClient := apiManagerClient()
	secClient := secretStoreClient()
	iamClient := identityManagerClient()

	deleteMap = map[string]ModelAction{
		utils.ImageKind:          CallDeleteImage(imgClient),
		utils.BaseImageKind:      CallDeleteBaseImage(imgClient),
		utils.FunctionKind:       CallDeleteFunction(fnClient),
		utils.SecretKind:         CallDeleteSecret(secClient),
		utils.PolicyKind:         CallDeletePolicy(iamClient),
		utils.ServiceAccountKind: CallDeleteServiceAccount(iamClient),
		utils.DriverTypeKind:     CallDeleteEventDriverType(eventClient),
		utils.DriverKind:         CallDeleteEventDriver(eventClient),
		utils.SubscriptionKind:   CallDeleteSubscription(eventClient),
		utils.APIKind:            CallDeleteAPI(apiClient),
	}
}

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

			initDeleteMap()
			if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
				isURL = true
				baseURL = file[:strings.LastIndex(file, "/")+1]
				err := importFileWithURL(out, errOut, cmd, args, deleteMap, "Deleted")
				CheckErr(err)
			} else {
				err := importFile(out, errOut, cmd, args, deleteMap, "Deleted")
				CheckErr(err)
			}
		},
		SuggestFor: []string{"list"},
	}
	cmd.AddCommand(NewCmdDeleteBaseImage(out, errOut))
	cmd.AddCommand(NewCmdDeleteImage(out, errOut))
	cmd.AddCommand(NewCmdDeleteSeedImage(out, errOut))
	cmd.AddCommand(NewCmdDeleteFunction(out, errOut))
	cmd.AddCommand(NewCmdDeleteSecret(out, errOut))
	cmd.AddCommand(NewCmdDeleteAPI(out, errOut))
	cmd.AddCommand(NewCmdDeleteSubscription(out, errOut))
	cmd.AddCommand(NewCmdDeleteEventDriver(out, errOut))
	cmd.AddCommand(NewCmdDeleteEventDriverType(out, errOut))

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to YAML file or an URL")
	cmd.Flags().StringVarP(&workDir, "work-dir", "w", "", "Working directory relative paths are based on")
	return cmd
}
