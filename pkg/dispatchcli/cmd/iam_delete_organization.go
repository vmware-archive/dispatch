///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/client"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteOrganizationLong = i18n.T(`Delete a dispatch organization`)

	// TODO: add examples
	deleteOrganizationExample = i18n.T(``)
)

// NewCmdIamDeleteOrganization creates command for delete service accounts
func NewCmdIamDeleteOrganization(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   i18n.T("organization ORGANIZATION_NAME"),
		Short: i18n.T("Delete organization"),
		Long:  deleteOrganizationLong,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := identityManagerClient()
			err := deleteOrganization(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeleteOrganization makes the API call to delete Organization
func CallDeleteOrganization(c client.IdentityClient) ModelAction {
	return func(s interface{}) error {
		organizationModel := s.(*v1.Organization)

		deleted, err := c.DeleteOrganization(context.TODO(), "", *organizationModel.Name)
		if err != nil {
			return err
		}
		*organizationModel = *deleted
		return nil
	}
}

func deleteOrganization(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.IdentityClient) error {
	organizationModel := v1.Organization{
		Name: &args[0],
	}

	err := CallDeleteOrganization(c)(&organizationModel)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(organizationModel)
	}
	fmt.Fprintf(out, "Deleted Organization: %s\n", *organizationModel.Name)
	return nil
}
