///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/cmd/utils"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/subscriptions"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

var (
	getSubscriptionsLong = i18n.T(`Get subscriptions.`)

	// TODO: add examples
	getSubscriptionsExample = i18n.T(``)
)

// NewCmdGetSubscription creates command responsible for getting subscriptions.
func NewCmdGetSubscription(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription [SUBSCRIPTION]",
		Short:   i18n.T("Get subscriptions"),
		Long:    getSubscriptionsLong,
		Example: getSubscriptionsExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"subscriptions"},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if len(args) > 0 {
				err = getSubscription(out, errOut, cmd, args)
			} else {
				err = getSubscriptions(out, errOut, cmd)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

func getSubscription(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	client := eventManagerClient()
	params := &subscriptions.GetSubscriptionParams{
		Context:          context.Background(),
		SubscriptionName: args[0],
		Tags:             []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.Subscriptions.GetSubscription(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatSubscriptionOutput(out, false, []*models.Subscription{resp.Payload})
}

func getSubscriptions(out, errOut io.Writer, cmd *cobra.Command) error {
	client := eventManagerClient()
	params := &subscriptions.GetSubscriptionsParams{
		Context: context.Background(),
		Tags:    []string{},
	}
	utils.AppendApplication(&params.Tags, cmdFlagApplication)

	resp, err := client.Subscriptions.GetSubscriptions(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	return formatSubscriptionOutput(out, true, resp.Payload)
}

func formatSubscriptionOutput(out io.Writer, list bool, subscriptions []*models.Subscription) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(subscriptions)
		}
		return encoder.Encode(subscriptions[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Source Type", "Source Name", "Event Type", "Function name", "Status", "Created Date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, sub := range subscriptions {
		table.Append([]string{*sub.Name, *sub.SourceType, *sub.SourceName, *sub.EventType, *sub.Function, string(sub.Status), time.Unix(sub.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
