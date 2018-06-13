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
	deleteServiceAccountLong = i18n.T(`Delete a dispatch service account`)

	// TODO: add examples
	deleteServiceAccountExample = i18n.T(``)
)

// NewCmdIamDeleteServiceAccount creates command for delete service accounts
func NewCmdIamDeleteServiceAccount(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   i18n.T("serviceaccount SERVICE_ACCOUNT_NAME"),
		Short: i18n.T("Delete serviceaccount"),
		Long:  deleteServiceAccountLong,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := identityManagerClient()
			err := deleteServiceAccount(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeleteServiceAccount makes the API call to delete ServiceAccount
func CallDeleteServiceAccount(c client.IdentityClient) ModelAction {
	return func(s interface{}) error {
		serviceAccountModel := s.(*v1.ServiceAccount)

		deleted, err := c.DeleteServiceAccount(context.TODO(), "", *serviceAccountModel.Name)
		if err != nil {
			return err
		}
		*serviceAccountModel = *deleted
		return nil
	}
}

func deleteServiceAccount(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.IdentityClient) error {
	serviceAccountModel := v1.ServiceAccount{
		Name: &args[0],
	}

	err := CallDeleteServiceAccount(c)(&serviceAccountModel)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(serviceAccountModel)
	}
	fmt.Fprintf(out, "Deleted ServiceAccount: %s\n", *serviceAccountModel.Name)
	return nil
}
