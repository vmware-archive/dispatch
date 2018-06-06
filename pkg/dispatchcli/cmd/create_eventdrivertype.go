///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
)

var (
	createEventDriverTypeLong = i18n.T(``)
	// TODO: add examples
	createEventDriverTypeExample = i18n.T(``)
)

// NewCmdCreateEventDriverType creates command responsible for dispatch function eventDriver creation.
func NewCmdCreateEventDriverType(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eventdrivertype DRIVER_TYPE_NAME DOCKER_IMAGE",
		Short:   i18n.T("Create an event driver type based on docker image."),
		Long:    createEventDriverTypeLong,
		Example: createEventDriverTypeExample,
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"eventdrivertypes", "event-driver-type", "event-driver-types", "eventdriver-types", "eventdriver-type"},
		Run: func(cmd *cobra.Command, args []string) {
			err := createEventDriverType(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	return cmd
}

// CallCreateEventDriverType makes the API call to create an event driver type
func CallCreateEventDriverType(i interface{}) error {
	client := eventManagerClient()
	driverTypeModel := i.(*v1.EventDriverType)

	params := &drivers.AddDriverTypeParams{
		Body:    driverTypeModel,
		Context: context.Background(),
	}

	created, err := client.Drivers.AddDriverType(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*driverTypeModel = *created.Payload
	return nil
}

func createEventDriverType(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	typeName := args[0]
	image := args[1]

	eventDriverType := &v1.EventDriverType{
		Name:  swag.String(typeName),
		Image: swag.String(image),
		Tags:  []*v1.Tag{},
	}
	if cmdFlagApplication != "" {
		eventDriverType.Tags = append(eventDriverType.Tags, &v1.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	err := CallCreateEventDriverType(eventDriverType)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(eventDriverType)
	}
	fmt.Fprintf(out, "Created event driver type: %s\n", *eventDriverType.Name)
	return nil
}
