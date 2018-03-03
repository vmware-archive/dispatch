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

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	client "github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
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
		Aliases: []string{"eventdrivertypes", "event-driver-type", "event-driver-types"},
		Run: func(cmd *cobra.Command, args []string) {
			err := createEventDriverType(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	return cmd
}

func createEventDriverType(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	typeName := args[0]
	image := args[1]

	eventDriverType := &models.DriverType{
		Name:  swag.String(typeName),
		Image: swag.String(image),
		Tags:  []*models.Tag{},
	}
	if cmdFlagApplication != "" {
		eventDriverType.Tags = append(eventDriverType.Tags, &models.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	params := &client.AddDriverTypeParams{
		Body:    eventDriverType,
		Context: context.Background(),
	}
	client := eventManagerClient()

	created, err := client.Drivers.AddDriverType(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Fprintf(out, "Created event driver type: %s\n", *created.Payload.Name)
	return nil
}
