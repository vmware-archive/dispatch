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

	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	client "github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
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

func deleteEventDriverType(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	params := &client.DeleteDriverTypeParams{
		Context:        context.Background(),
		DriverTypeName: args[0],
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := eventManagerClient().Drivers.DeleteDriverType(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatDeleteEventDriverTypeOutput(out, false, []*models.DriverType{resp.Payload})
}

func formatDeleteEventDriverTypeOutput(out io.Writer, list bool, driverTypes []*models.DriverType) error {
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
