///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	serviceinstance "github.com/vmware/dispatch/pkg/service-manager/gen/client/service_instance"
	models "github.com/vmware/dispatch/pkg/service-manager/gen/models"
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
			if len(args) == 1 {
				err = getServiceInstance(out, errOut, cmd, args)
			} else {
				err = getServiceInstances(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getServiceInstance(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := serviceManagerClient()
	params := &serviceinstance.GetServiceInstanceByNameParams{
		Context:             context.Background(),
		ServiceInstanceName: args[0],
	}

	resp, err := client.ServiceInstance.GetServiceInstanceByName(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	if resp.Payload.Name == nil {
		err := serviceinstance.NewGetServiceInstanceByNameNotFound()
		err.Payload = &models.Error{
			Code:    404,
			Message: &args[0],
		}
		return formatAPIError(err, params)
	}

	return formatServiceInstanceOutput(out, false, []*models.ServiceInstance{resp.Payload})
}

func getServiceInstances(out, errOut io.Writer, cmd *cobra.Command) error {
	client := serviceManagerClient()
	params := &serviceinstance.GetServiceInstancesParams{
		Context: context.Background(),
		Tags:    []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.ServiceInstance.GetServiceInstances(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatServiceInstanceOutput(out, true, resp.Payload)
}

func formatServiceInstanceOutput(out io.Writer, list bool, serviceInstances []*models.ServiceInstance) error {

	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(serviceInstances)
		}
		return encoder.Encode(serviceInstances[0])
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Class", "Plan", "Provisioned", "Bound", "Status"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, instance := range serviceInstances {
		provisioned := instance.Status == models.StatusREADY
		bound := false
		if instance.Binding != nil {
			bound = instance.Binding.Status == models.StatusREADY
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
