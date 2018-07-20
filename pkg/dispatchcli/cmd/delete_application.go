///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/application-manager/gen/client/application"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteApplicationsLong = i18n.T(`Delete applications.`)

	// TODO: add examples
	deleteApplicationsExample = i18n.T(``)
)

// NewCmdDeleteApplication creates command responsible for deleting  applications.
func NewCmdDeleteApplication(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application NAME",
		Short:   i18n.T("Delete application"),
		Long:    deleteApplicationsLong,
		Example: deleteApplicationsExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"app"},
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteApplication(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeleteApplication makes the API call to delete an application
func CallDeleteApplication(i interface{}) error {
	client := applicationManagerClient()
	applicationModel := i.(*v1.Application)
	params := &application.DeleteAppParams{
		Application:  *applicationModel.Name,
		Context:      context.Background(),
		XDispatchOrg: getOrgFromConfig(),
	}

	deleted, err := client.Application.DeleteApp(params, GetAuthInfoWriter())
	if err != nil {
		return err
	}
	*applicationModel = *deleted.Payload
	return nil
}

func deleteApplication(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	applicationModel := v1.Application{
		Name: &args[0],
	}
	err := CallDeleteApplication(&applicationModel)
	if err != nil {
		return err
	}
	return formatDeleteApplicationOutput(out, false, []*v1.Application{&applicationModel})
}

func formatDeleteApplicationOutput(out io.Writer, list bool, applications []*v1.Application) error {
	if w, err := formatOutput(out, list, applications); w {
		return err
	}
	for _, i := range applications {
		_, err := fmt.Fprintf(out, "Deleted application: %s\n", *i.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
