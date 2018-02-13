///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/application-manager/gen/client/application"
	"github.com/vmware/dispatch/pkg/application-manager/gen/models"
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
	applicationModel := i.(*models.Application)
	params := &application.DeleteAppParams{
		Application: *applicationModel.Name,
		Context:     context.Background(),
	}

	deleted, err := client.Application.DeleteApp(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*applicationModel = *deleted.Payload
	return nil
}

func deleteApplication(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	applicationModel := models.Application{
		Name: &args[0],
	}
	err := CallDeleteApplication(&applicationModel)
	if err != nil {
		return err
	}
	return formatDeleteApplicationOutput(out, false, []*models.Application{&applicationModel})
}

func formatDeleteApplicationOutput(out io.Writer, list bool, applications []*models.Application) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(applications)
		}
		return encoder.Encode(applications[0])
	}
	for _, i := range applications {
		_, err := fmt.Fprintf(out, "Deleted application: %s\n", *i.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
