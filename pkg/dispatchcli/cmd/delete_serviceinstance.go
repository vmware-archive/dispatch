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

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	serviceinstance "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_instance"
	models "github.com/vmware/dispatch/pkg/service-manager/gen/models"
)

var (
	deleteServiceInstanceLong = i18n.T(`Delete service instance.`)

	// TODO: add examples
	deleteServiceInstanceExample = i18n.T(``)
)

// NewCmdDeleteServiceInstance creates command responsible for deleting a service instance
func NewCmdDeleteServiceInstance(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceinstance SERVICE_INSTANCE_NAME",
		Short:   i18n.T("Delete service instance"),
		Long:    deleteServiceInstanceLong,
		Example: deleteServiceInstanceExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"apis"},
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteServiceInstance(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteServiceInstance makes the API call to create an image
func CallDeleteServiceInstance(i interface{}) error {
	client := serviceManagerClient()
	serviceInstanceModel := i.(*models.ServiceInstance)
	params := &serviceinstance.DeleteServiceInstanceByNameParams{
		ServiceInstanceName: *serviceInstanceModel.Name,
		Context:             context.Background(),
	}
	deleted, err := client.ServiceInstance.DeleteServiceInstanceByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*serviceInstanceModel = *deleted.Payload
	return nil
}

func deleteServiceInstance(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	serviceInstanceModel := models.ServiceInstance{
		Name: &args[0],
	}
	err := CallDeleteServiceInstance(&serviceInstanceModel)
	if err != nil {
		return err
	}
	return formatDeleteServiceInstanceOutput(out, false, []*models.ServiceInstance{&serviceInstanceModel})
}

func formatDeleteServiceInstanceOutput(out io.Writer, list bool, services []*models.ServiceInstance) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(services)
		}
		return encoder.Encode(services[0])
	}
	for _, s := range services {
		_, err := fmt.Fprintf(out, "Deleted service instance: %s\n", *s.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
