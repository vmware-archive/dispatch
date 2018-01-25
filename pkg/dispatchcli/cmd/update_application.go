///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/application-manager/gen/client/application"
	"github.com/vmware/dispatch/pkg/application-manager/gen/models"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	updateApplicationLong = i18n.T(`Update a dispatch application.
		APPLICATION_FILE - the path to a .json file describing an application

		Example: application.json:
		{
			"application-key": "secret-value"
		}`)

	// TODO: add examples
	updateApplicationExample = i18n.T(`update a application`)
)

// CallUpdateApplication makes the API call to update an application
func CallUpdateApplication(input interface{}) error {
	client := applicationManagerClient()
	applicationBody := input.(*models.Application)

	params := application.NewUpdateAppParams()
	params.Application = *applicationBody.Name
	params.Body = applicationBody

	_, err := client.Application.UpdateApp(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return err
}

// NewCmdUpdateApplication updates a dispatch application
func NewCmdUpdateApplication(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Short:   i18n.T("Update application"),
		Long:    updateApplicationLong,
		Example: createApplicationExample,
		Args:    cobra.MinimumNArgs(0),
		Aliases: []string{"app"},
		Run: func(cmd *cobra.Command, args []string) {
			err := updateApplication(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

func updateApplication(out io.Writer, errOut io.Writer, cmd *cobra.Command, args []string) error {

	fmt.Printf("Update Application not yet implemented")
	return nil
}
