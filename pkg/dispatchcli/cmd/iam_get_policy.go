///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	policy "github.com/vmware/dispatch/pkg/identity-manager/gen/client/policy"

	models "github.com/vmware/dispatch/pkg/identity-manager/gen/models"
	"golang.org/x/net/context"
)

var (
	getPoliciesLong = i18n.T(`Get policies`)

	// TODO: examples
	getPoliciesExample = i18n.T(``)

	printRuleContent = false
)

// NewCmdIamGetPolicy creates command for getting policies
func NewCmdIamGetPolicy(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     i18n.T("policy [POLICY_NAME]"),
		Short:   i18n.T("Get policies"),
		Long:    getPoliciesLong,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"policies"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) > 0 {
				err = getPolicy(out, errOut, cmd, args)
			} else {
				err = getPolicies(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().BoolVarP(&printRuleContent, "wide", "w", false, "print rule context")
	return cmd
}

func getPolicy(out, errOut io.Writer, cmd *cobra.Command, args []string) error {

	client := identityManagerClient()
	params := &policy.GetPolicyParams{
		PolicyName: args[0],
		Context:    context.Background(),
	}

	resp, err := client.Policy.GetPolicy(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	if resp.Payload.Name == nil {
		err := policy.NewGetPolicyNotFound()
		err.Payload = &models.Error{
			Code:    404,
			Message: &args[0],
		}
		return formatAPIError(err, params)
	}

	return formatPolicyOutput(out, false, []*models.Policy{resp.Payload})
}

func getPolicies(out, errOut io.Writer, cmd *cobra.Command) error {

	client := identityManagerClient()
	params := &policy.GetPoliciesParams{
		Context: context.Background(),
	}

	resp, err := client.Policy.GetPolicies(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatPolicyOutput(out, true, resp.Payload)
}

func formatPolicyOutput(out io.Writer, list bool, policies []*models.Policy) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(policies)
		}
		return encoder.Encode(policies[0])
	}

	headers := []string{"Name", "Created Date"}
	if printRuleContent {
		headers = append(headers, "Rules")
	} else {
		fmt.Fprintf(out, "Note: rule contents are omitted, please use --wide or -w to print them\n")
	}

	table := tablewriter.NewWriter(out)
	table.SetHeader(headers)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	table.SetAutoWrapText(false)
	for _, policy := range policies {
		// For now, a policy has one rule
		ruleContent, err := json.MarshalIndent(policy.Rules[0], "", "  ")
		row := []string{*policy.Name, time.Unix(policy.CreatedTime, 0).Local().Format(time.UnixDate)}
		if printRuleContent && err == nil {
			row = append(row, string(ruleContent))
		}
		table.Append(row)
	}
	table.Render()
	return nil
}
