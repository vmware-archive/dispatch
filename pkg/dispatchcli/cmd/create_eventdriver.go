///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	client "github.com/vmware/dispatch/pkg/event-manager/gen/client/drivers"
)

var (
	createEventDriverLong = i18n.T(
		`Create dispatch event driver

Types and Settings:
* vcenter
	- vcenterurl 	string (required) (e.g. <user>:<password>@<vcenter-host>:<vcenter-port> )
		`)
	// TODO: add examples
	createEventDriverExample = i18n.T(``)

	createEventDriverConfig  []string
	createEventDriverSecrets []string
	createEventDriverName    string
)

// NewCmdCreateEventDriver creates command responsible for dispatch function eventDriver creation.
func NewCmdCreateEventDriver(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eventdriver DRIVER_TYPE [--name DRIVER_NAME] [--set KEY=VALUE] [--secret SECRET_NAME]",
		Short:   i18n.T("Create event driver"),
		Long:    createEventDriverLong,
		Example: createEventDriverExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"eventdrivers", "event-driver", "event-drivers"},
		Run: func(cmd *cobra.Command, args []string) {
			err := createEventDriver(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&createEventDriverName, "name", "", "name for the event driver. will be automatically generated if not specified.")
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringArrayVarP(&createEventDriverConfig, "set", "s", []string{}, "set event driver configurations, default: empty")
	cmd.Flags().StringArrayVar(&createEventDriverSecrets, "secret", []string{}, "Configuration passed via secrets, can be specified multiple times or a comma-delimited string")
	return cmd
}

func createEventDriver(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	driverType := args[0]

	var driverConfig []*v1.Config
	for _, conf := range createEventDriverConfig {
		result := strings.Split(conf, "=")
		switch len(result) {
		case 1:
			driverConfig = append(driverConfig, &v1.Config{
				Key: result[0],
			})
		case 2:
			driverConfig = append(driverConfig, &v1.Config{
				Key:   result[0],
				Value: result[1],
			})
		default:
			fmt.Fprint(errOut, "Invalid Configuration Format, should be --set key=value or --set key")
		}
	}

	eventDriver := &v1.EventDriver{
		Name:    swag.String(resourceName(createEventDriverName)),
		Type:    swag.String(driverType),
		Config:  driverConfig,
		Secrets: createEventDriverSecrets,
		Tags:    []*v1.Tag{},
	}
	if cmdFlagApplication != "" {
		eventDriver.Tags = append(eventDriver.Tags, &v1.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}

	params := &client.AddDriverParams{
		Body:    eventDriver,
		Context: context.Background(),
	}

	created, err := eventManagerClient().Drivers.AddDriver(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Fprintf(out, "Created event driver: %s\n", *created.Payload.Name)
	return nil
}
