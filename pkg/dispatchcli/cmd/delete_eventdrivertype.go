///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteEventDriverTypeLong = i18n.T(`Delete event driver type`)

	// TODO: add examples
	deleteEventDriverTypeExample = i18n.T(``)
)

// NewCmdDeleteEventDriverType creates command responsible for deleting EventDriverType.
func NewCmdDeleteEventDriverType(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eventdrivertype TYPE_NAME",
		Short:   i18n.T("Delete event driver type"),
		Long:    deleteEventDriverTypeLong,
		Example: deleteEventDriverTypeExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"eventdrivertypes", "event-driver-type", "event-driver-types", "eventdriver-types", "eventdriver-type"},
		Run: func(cmd *cobra.Command, args []string) {
			c := eventManagerClient()
			err := deleteEventDriverType(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteEventDriverType makes the API call to delete an event driver
func CallDeleteEventDriverType(c client.EventsClient) ModelAction {
	return func(i interface{}) error {
		driverType := i.(*v1.EventDriverType)

		deleted, err := c.DeleteEventDriverType(context.TODO(), "", *driverType.Name)
		if err != nil {
			return err
		}
		*driverType = *deleted
		return nil
	}
}

func deleteEventDriverType(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EventsClient) error {

	driverTypeModel := v1.EventDriverType{
		Name: &args[0],
	}
	err := CallDeleteEventDriverType(c)(&driverTypeModel)
	if err != nil {
		return err
	}
	return formatDeleteEventDriverTypeOutput(out, false, []*v1.EventDriverType{&driverTypeModel})
}

func formatDeleteEventDriverTypeOutput(out io.Writer, list bool, driverTypes []*v1.EventDriverType) error {
	if w, err := formatOutput(out, list, driverTypes); w {
		return err
	}
	for _, d := range driverTypes {
		_, err := fmt.Fprintf(out, "Deleted driver types: %s\n", *d.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
