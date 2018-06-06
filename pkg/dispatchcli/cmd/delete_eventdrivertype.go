///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/vmware/dispatch/pkg/api/v1"
	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
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
			err := deleteEventDriverType(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteEventDriverType makes the API call to delete an event driver
func CallDeleteEventDriverType(i interface{}) error {
	client := eventManagerClient()
	driverTypeModel := i.(*v1.EventDriverType)
	params := &drivers.DeleteDriverTypeParams{
		Context:        context.Background(),
		DriverTypeName: *driverTypeModel.Name,
		Tags:           []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	deleted, err := client.Drivers.DeleteDriverType(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*driverTypeModel = *deleted.Payload
	return nil
}

func deleteEventDriverType(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	driverTypeModel := v1.EventDriverType{
		Name: &args[0],
	}
	err := CallDeleteEventDriverType(&driverTypeModel)
	if err != nil {
		return err
	}
	return formatDeleteEventDriverTypeOutput(out, false, []*v1.EventDriverType{&driverTypeModel})
}

func formatDeleteEventDriverTypeOutput(out io.Writer, list bool, driverTypes []*v1.EventDriverType) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(driverTypes)
		}
		return encoder.Encode(driverTypes[0])
	}
	for _, d := range driverTypes {
		_, err := fmt.Fprintf(out, "Deleted driver types: %s\n", *d.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
