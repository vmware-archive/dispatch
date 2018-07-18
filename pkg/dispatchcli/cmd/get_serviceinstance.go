///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"golang.org/x/net/context"
)

var (
	getServiceInstancesLong = i18n.T(`Get service instances.`)

	// TODO: add examples
	getServiceInstancesExample = i18n.T(``)
)

// NewCmdGetServiceInstance creates command responsible for getting service instances.
func NewCmdGetServiceInstance(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceinstance [SERVICE_INSTANCE_NAME ...]",
		Short:   i18n.T("Get serviceinstances"),
		Long:    getServiceInstancesLong,
		Example: getServiceInstancesExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"serviceinstances"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := serviceManagerClient()
			if len(args) == 1 {
				err = getServiceInstance(out, errOut, cmd, args, c)
			} else {
				err = getServiceInstances(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getServiceInstance(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ServicesClient) error {
	serviceInstanceName := args[0]

	resp, err := c.GetServiceInstance(context.TODO(), dispatchConfig.Organization, serviceInstanceName)
	if err != nil {
		return err
	}
	return formatServiceInstanceOutput(out, false, []v1.ServiceInstance{*resp})
}

func getServiceInstances(out, errOut io.Writer, cmd *cobra.Command, c client.ServicesClient) error {
	resp, err := c.ListServiceInstances(context.TODO(), dispatchConfig.Organization)
	if err != nil {
		return err
	}
	return formatServiceInstanceOutput(out, true, resp)
}

func formatServiceInstanceOutput(out io.Writer, list bool, serviceInstances []v1.ServiceInstance) error {

	if w, err := formatOutput(out, list, serviceInstances); w {
		return err
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Class", "Plan", "Provisioned", "Bound", "Status"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, instance := range serviceInstances {
		provisioned := instance.Status == v1.StatusREADY
		bound := false
		if instance.Binding != nil {
			bound = instance.Binding.Status == v1.StatusREADY
		}

		status := instance.Status
		if provisioned && (instance.Binding != nil && !bound) {
			status = instance.Binding.Status
		}

		table.Append([]string{*instance.Name, *instance.ServiceClass, *instance.ServicePlan, strconv.FormatBool(provisioned), strconv.FormatBool(bound), string(status)})
	}
	table.Render()
	return nil
}
