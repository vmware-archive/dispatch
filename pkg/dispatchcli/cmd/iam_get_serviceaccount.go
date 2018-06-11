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

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/identity-manager/gen/client/serviceaccount"
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
			if len(args) > 0 {
				err = getServiceAccount(out, errOut, cmd, args)
			} else {
				err = getServiceAccounts(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getServiceAccount(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	client := identityManagerClient()
	params := &serviceaccount.GetServiceAccountParams{
		ServiceAccountName: args[0],
		Context:            context.Background(),
		XDispatchOrg:       getOrganization(),
	}

	resp, err := client.Serviceaccount.GetServiceAccount(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	if resp.Payload.Name == nil {
		err := serviceaccount.NewGetServiceAccountNotFound()
		err.Payload = &v1.Error{
			Code:    404,
			Message: &args[0],
		}
		return formatAPIError(err, params)
	}

	return formatServiceAccountOutput(out, false, []*v1.ServiceAccount{resp.Payload})
}

func getServiceAccounts(out, errOut io.Writer, cmd *cobra.Command) error {

	client := identityManagerClient()
	params := &serviceaccount.GetServiceAccountsParams{
		Context:      context.Background(),
		XDispatchOrg: getOrganization(),
	}

	resp, err := client.Serviceaccount.GetServiceAccounts(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatServiceAccountOutput(out, true, resp.Payload)
}

func formatServiceAccountOutput(out io.Writer, list bool, serviceAccounts []*v1.ServiceAccount) error {

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
