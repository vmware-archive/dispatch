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
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
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
		Aliases: []string{"serviceinstances"},
		Run: func(cmd *cobra.Command, args []string) {
			c := serviceManagerClient()
			err := deleteServiceInstance(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteServiceInstance makes the API call to create an image
func CallDeleteServiceInstance(c client.ServicesClient) ModelAction {
	return func(i interface{}) error {
		serviceInstanceModel := i.(*v1.ServiceInstance)

		err := c.DeleteServiceInstance(context.TODO(), dispatchConfig.Organization, *serviceInstanceModel.Name)
		if err != nil {
			return err
		}
		return nil
	}
}

func deleteServiceInstance(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ServicesClient) error {
	serviceInstanceModel := v1.ServiceInstance{
		Name: &args[0],
	}
	err := CallDeleteServiceInstance(c)(&serviceInstanceModel)
	if err != nil {
		return err
	}
	return formatDeleteServiceInstanceOutput(out, false, []*v1.ServiceInstance{&serviceInstanceModel})
}

func formatDeleteServiceInstanceOutput(out io.Writer, list bool, services []*v1.ServiceInstance) error {
	if w, err := formatOutput(out, list, services); w {
		return err
	}
	for _, s := range services {
		_, err := fmt.Fprintf(out, "Deleted service instance: %s\n", *s.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
