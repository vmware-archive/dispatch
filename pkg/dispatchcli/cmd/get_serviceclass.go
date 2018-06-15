///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"golang.org/x/net/context"
)

var (
	getServiceClassesLong = i18n.T(`Get service classes.`)

	// TODO: add examples
	getServiceClassesExample = i18n.T(``)
)

// NewCmdGetServiceClass creates command responsible for getting service classes.
func NewCmdGetServiceClass(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceclass [SERVICE_CLASS_NAME ...]",
		Short:   i18n.T("Get serviceclasses"),
		Long:    getServiceClassesLong,
		Example: getServiceClassesExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"serviceclasses"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := serviceManagerClient()
			if len(args) == 1 {
				err = getServiceClass(out, errOut, cmd, args, c)
			} else {
				err = getServiceClasses(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getServiceClass(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.ServicesClient) error {
	serviceClassName := args[0]

	resp, err := c.GetServiceClass(context.TODO(), serviceClassName)
	if err != nil {
		return err
	}
	return formatServiceClassOutput(out, false, []v1.ServiceClass{*resp})
}

func getServiceClasses(out, errOut io.Writer, cmd *cobra.Command, c client.ServicesClient) error {
	resp, err := c.ListServiceClasses(context.TODO())
	if err != nil {
		return err
	}
	return formatServiceClassOutput(out, true, resp)
}

func formatServiceClassOutput(out io.Writer, list bool, serviceClasses []v1.ServiceClass) error {

	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(serviceClasses)
		}
		return encoder.Encode(serviceClasses[0])
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Broker", "Bindable", "Plans", "Status"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, class := range serviceClasses {
		var plans []string
		for _, plan := range class.Plans {
			plans = append(plans, plan.Name)
		}
		table.Append([]string{*class.Name, *class.Broker, strconv.FormatBool(class.Bindable), strings.Join(plans, "\n"), string(class.Status)})
	}
	table.Render()
	return nil
}
