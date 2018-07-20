///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/client"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createOrganizationLong = i18n.T(`Create a dispatch organization`)

	createOrganizationExample = i18n.T(`
# Create a organization
dispatch iam create organization <organization_name>
`)
)

// NewCmdIamCreateOrganization creates command responsible for org creation
func NewCmdIamCreateOrganization(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T(`organization ORGANIZATION_NAME`),
		Short:   i18n.T(`Create organization`),
		Long:    createOrganizationLong,
		Example: createOrganizationExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := identityManagerClient()
			err := createOrganization(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	return cmd
}

// CallCreateOrganization makes the api call to create a organization
func callCreateOrganization(c client.IdentityClient) ModelAction {
	return func(p interface{}) error {
		organizationModel := p.(*v1.Organization)

		created, err := c.CreateOrganization(context.TODO(), "", organizationModel)
		if err != nil {
			return err
		}

		*organizationModel = *created
		return nil
	}
}

func createOrganization(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.IdentityClient) error {
	organizationName := args[0]

	organizationModel := &v1.Organization{
		Name: &organizationName,
	}

	err := callCreateOrganization(c)(organizationModel)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, organizationModel); w {
		return err
	}
	fmt.Fprintf(out, "Created organization: %s\n", *organizationModel.Name)
	return nil
}
