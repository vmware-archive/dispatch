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
	"github.com/vmware/dispatch/pkg/client"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
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
			c := identityManagerClient()
			if len(args) > 0 {
				err = getPolicy(out, errOut, cmd, args, c)
			} else {
				err = getPolicies(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().BoolVarP(&printRuleContent, "wide", "w", false, "print rule context")
	return cmd
}

func getPolicy(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.IdentityClient) error {
	resp, err := c.GetPolicy(context.TODO(), "", args[0])
	if err != nil {
		return err
	}

	return formatPolicyOutput(out, false, []v1.Policy{*resp})
}

func getPolicies(out, errOut io.Writer, cmd *cobra.Command, c client.IdentityClient) error {
	resp, err := c.ListPolicies(context.TODO(), "")
	if err != nil {
		return err
	}
	return formatPolicyOutput(out, true, resp)
}

func formatPolicyOutput(out io.Writer, list bool, policies []v1.Policy) error {
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
