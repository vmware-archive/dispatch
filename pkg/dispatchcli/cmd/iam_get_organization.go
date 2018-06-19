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
	getOrganizationsLong = i18n.T(`Get organizations`)

	// TODO: examples
	getOrganizationsExample = i18n.T(``)
)

// NewCmdIamGetOrganization creates command for getting organizations
func NewCmdIamGetOrganization(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T("organization [ORGANIZATION_NAME]"),
		Short:   i18n.T("Get organizations"),
		Long:    getOrganizationsLong,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"organizations"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			c := identityManagerClient()
			if len(args) > 0 {
				err = getOrganization(out, errOut, cmd, args, c)
			} else {
				err = getOrganizations(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	return cmd
}

func getOrganization(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.IdentityClient) error {
	resp, err := c.GetOrganization(context.TODO(), "", args[0])
	if err != nil {
		return err
	}

	return formatOrganizationOutput(out, false, []v1.Organization{*resp})
}

func getOrganizations(out, errOut io.Writer, cmd *cobra.Command, c client.IdentityClient) error {
	resp, err := c.ListOrganizations(context.TODO(), "")
	if err != nil {
		return err
	}
	return formatOrganizationOutput(out, true, resp)
}

func formatOrganizationOutput(out io.Writer, list bool, organizations []v1.Organization) error {

	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(organizations)
		}
		return encoder.Encode(organizations[0])
	}

	headers := []string{"Name", "Created Date"}
	table := tablewriter.NewWriter(out)
	table.SetHeader(headers)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, organization := range organizations {
		row := []string{*organization.Name, time.Unix(organization.CreatedTime, 0).Local().Format(time.UnixDate)}
		table.Append(row)
	}
	table.Render()
	return nil
}
