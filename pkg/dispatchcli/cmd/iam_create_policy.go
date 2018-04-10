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
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"

	policy "github.com/vmware/dispatch/pkg/identity-manager/gen/client/policy"
	models "github.com/vmware/dispatch/pkg/identity-manager/gen/models"
)

var (
	createPolicyLong = i18n.T(`Create a dispatch policy`)

	createPolicyExample = i18n.T(`
# Create a policy by specifying a single subject, action, and resource
dispatch iam create policy example_policy --subject user1@example.com --action get --resource "*"

# Create a policy by specifying multiple subjects, actions, and resources
dispatch iam create policy example_policy --subject user1@example.com,user2@example.com --action get,create,delete,update --resource image,function,base-image,secret

dispatch iam create policy example_policy --subject user1@example.com --subject user2@example.com --action get --action create,delete,update --resource image,function --resource base-image,secret
`)

	subjects  *[]string
	actions   *[]string
	resources *[]string
)

// NewCmdIamCreatePolicy creates command responsible for dispatch policy creation
func NewCmdIamCreatePolicy(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T(`policy POLICY_NAME --subject SUBJECTS --action ACTIONS --resource RESOURCES`),
		Short:   i18n.T("Create policy"),
		Long:    createPolicyLong,
		Example: createPolicyExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := createPolicy(out, errOut, cmd, args)
			CheckErr(err)
		},
	}

	subjects = cmd.Flags().StringSliceP("subject", "s", []string{""}, "subjects of policy rule, separated by comma")
	actions = cmd.Flags().StringSliceP("action", "a", []string{""}, "actions of policy rule, separated by comma")
	resources = cmd.Flags().StringSliceP("resource", "r", []string{""}, "resources of policy rule, separated by comma")
	return cmd
}

// CallCreatePolicy makes the api call to create a policy
func CallCreatePolicy(p interface{}) error {
	client := identityManagerClient()
	policyModel := p.(*models.Policy)

	params := &policy.AddPolicyParams{
		Body:    policyModel,
		Context: context.Background(),
	}

	created, err := client.Policy.AddPolicy(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*policyModel = *created.Payload
	return nil
}

func createPolicy(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	policyName := args[0]
	policyRules := []*models.Rule{
		{
			Subjects:  *subjects,
			Actions:   *actions,
			Resources: *resources,
		},
	}

	policyModel := &models.Policy{
		Name:  &policyName,
		Rules: policyRules,
	}

	err := CallCreatePolicy(policyModel)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(policyModel)
	}
	fmt.Fprintf(out, "Created policy: %s\n", *policyModel.Name)
	return nil
}
