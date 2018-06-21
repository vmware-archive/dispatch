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
	"github.com/vmware/dispatch/pkg/client"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
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
			c := eventManagerClient()
			if len(args) > 0 {
				err = getSubscription(out, errOut, cmd, args, c)
			} else {
				err = getSubscriptions(out, errOut, cmd, c)
			}
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

func getSubscription(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EventsClient) error {
	subName := args[0]
	resp, err := c.GetSubscription(context.TODO(), "", subName)
	if err != nil {
		return err
	}
	return formatSubscriptionOutput(out, false, []v1.Subscription{*resp})
}

func getSubscriptions(out, errOut io.Writer, cmd *cobra.Command, c client.EventsClient) error {
	resp, err := c.ListSubscriptions(context.TODO(), "")
	if err != nil {
		return err
	}
	return formatSubscriptionOutput(out, true, resp)
}

func formatSubscriptionOutput(out io.Writer, list bool, subscriptions []v1.Subscription) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(subscriptions)
		}
		return encoder.Encode(subscriptions[0])
	}
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Name", "Event type", "Function name", "Status", "Created date"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("")
	for _, sub := range subscriptions {
		table.Append([]string{*sub.Name, *sub.EventType, *sub.Function, string(sub.Status), time.Unix(sub.CreatedTime, 0).Local().Format(time.UnixDate)})
	}
	table.Render()
	return nil
}
