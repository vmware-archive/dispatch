///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/client"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createEventDriverTypeLong = i18n.T(``)
	// TODO: add examples
	createEventDriverTypeExample = i18n.T(``)
	exposeEventDriverType        = false
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
			c := eventManagerClient()
			err := createEventDriverType(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().BoolVar(&exposeEventDriverType, "expose", false, "expose the driver externally")
	return cmd
}

// CallCreateEventDriverType makes the API call to create an event driver type
func CallCreateEventDriverType(c client.EventsClient) ModelAction {
	return func(driver interface{}) error {
		evt := driver.(*v1.EventDriverType)

		created, err := c.CreateEventDriverType(context.TODO(), "", evt)
		if err != nil {
			return err
		}
		*evt = *created
		return nil
	}
}

func createEventDriverType(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EventsClient) error {

	typeName := args[0]
	image := args[1]

	eventDriverType := &v1.EventDriverType{
		Name:   swag.String(typeName),
		Image:  swag.String(image),
		Expose: exposeEventDriverType,
		Tags:   []*v1.Tag{},
	}
	if cmdFlagApplication != "" {
		eventDriverType.Tags = append(eventDriverType.Tags, &v1.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	err := CallCreateEventDriverType(c)(eventDriverType)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, eventDriverType); w {
		return err
	}
	fmt.Fprintf(out, "Created event driver type: %s\n", *eventDriverType.Name)
	return nil
}
