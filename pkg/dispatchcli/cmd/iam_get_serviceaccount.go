///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/client"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	getServiceAccountsLong = i18n.T(`Get service accounts`)

	// TODO: examples
	getServiceAccountsExample = i18n.T(``)
)

// NewCmdIamGetServiceAccount creates command for getting service accounts
func NewCmdIamGetServiceAccount(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T("serviceaccount [SERVICE_ACCOUNT_NAME]"),
		Short:   i18n.T("Get serviceaccounts"),
		Long:    getServiceAccountsLong,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"serviceaccounts"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := identityManagerClient()
			if len(args) > 0 {
				err = getServiceAccount(out, errOut, cmd, args, c)
			} else {
				err = getServiceAccounts(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getServiceAccount(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.IdentityClient) error {
	resp, err := c.GetServiceAccount(context.TODO(), args[0])
	if err != nil {
		return err
	}

	return formatServiceAccountOutput(out, false, []v1.ServiceAccount{*resp})
}

func getServiceAccounts(out, errOut io.Writer, cmd *cobra.Command, c client.IdentityClient) error {
	resp, err := c.ListServiceAccounts(context.TODO())
	if err != nil {
		return err
	}
	return formatServiceAccountOutput(out, true, resp)
}

func formatServiceAccountOutput(out io.Writer, list bool, serviceAccounts []v1.ServiceAccount) error {

	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(serviceAccounts)
		}
		return encoder.Encode(serviceAccounts[0])
	}

	headers := []string{"Name", "Created Date"}
	table := tablewriter.NewWriter(out)
	table.SetHeader(headers)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, serviceAccount := range serviceAccounts {
		row := []string{*serviceAccount.Name, time.Unix(serviceAccount.CreatedTime, 0).Local().Format(time.UnixDate)}
		table.Append(row)
	}
	table.Render()
	return nil
}
