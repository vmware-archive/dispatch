///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"golang.org/x/net/context"
)

var (
	deletePolicyLong = i18n.T(`Delete a dispatch policy`)

	// TODO: add examples
	deletePolicyExample = i18n.T(``)
)

// NewCmdIamDeletePolicy deletes policy
func NewCmdIamDeletePolicy(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T("policy POLICY_NAME"),
		Short:   i18n.T("Delete policy"),
		Long:    deletePolicyLong,
		Example: deletePolicyExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := identityManagerClient()
			err := deletePolicy(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeletePolicy makes the API call to delete policy
func CallDeletePolicy(c client.IdentityClient) ModelAction {
	return func(p interface{}) error {
		policyModel := p.(*v1.Policy)

		deleted, err := c.DeletePolicy(context.TODO(), "", *policyModel.Name)
		if err != nil {
			return err
		}
		*policyModel = *deleted
		return nil
	}
}

func deletePolicy(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.IdentityClient) error {
	policyModel := v1.Policy{
		Name: &args[0],
	}

	err := CallDeletePolicy(c)(&policyModel)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(policyModel)
	}
	fmt.Fprintf(out, "Deleted policy: %s\n", *policyModel.Name)
	return nil
}
