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
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"

	models "github.com/vmware/dispatch/pkg/identity-manager/gen/models"

	policy_client "github.com/vmware/dispatch/pkg/identity-manager/gen/client/policy"
)

var (
	updatePolicyLong = i18n.T(`Update a policy`)

	updatePolicyExample = i18n.T(`
	# Update a policy by specifying subjects, actions, or resources
	dispatch update policy example_policy --subject user1@example.com,user2@example.com --action get
	`)

	updatedSubjects  *[]string
	updatedActions   *[]string
	updatedResources *[]string
)

// NewCmdUpdatePolicy create command for updating policy
func NewCmdUpdatePolicy(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T("policy POLICY_NAME --subject SUBJECTS --action ACTIONS --resource RESOURCES"),
		Short:   i18n.T("Update policy"),
		Long:    updatePolicyLong,
		Example: updatePolicyExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := updatePolicy(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	updatedSubjects = cmd.Flags().StringSliceP("subject", "s", []string{""}, "subjects of policy rule, separated by comma")
	updatedActions = cmd.Flags().StringSliceP("action", "a", []string{""}, "actions of policy rule, separated by comma")
	updatedResources = cmd.Flags().StringSliceP("resource", "r", []string{""}, "resources of policy rule, separated by comma")

	return cmd
}

// CallUpdatePolicy updates a policy
func CallUpdatePolicy(p interface{}) error {

	policyModel := p.(*models.Policy)

	params := &policy_client.UpdatePolicyParams{
		PolicyName: *policyModel.Name,
		Body:       policyModel,
		Context:    context.Background(),
	}

	_, err := identityManagerClient().Policy.UpdatePolicy(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}

func updatePolicy(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	params := &policy_client.GetPolicyParams{
		PolicyName: args[0],
		Context:    context.Background(),
	}

	policyOk, err := identityManagerClient().Policy.GetPolicy(params, GetAuthInfoWriter())

	if err != nil {
		return formatAPIError(err, params)
	}

	policyModel := *policyOk.Payload
	policyModel.Rules = models.PolicyRules{
		&models.Rule{
			Subjects:  *updatedSubjects,
			Actions:   *updatedActions,
			Resources: *updatedResources,
		},
	}

	err = CallUpdatePolicy(&policyModel)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated policy: %s\n", *policyModel.Name)
	return nil
}
