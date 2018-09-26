///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/vmware/dispatch/pkg/client"
	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	deleteSubscriptionLong = i18n.T(`Delete subscriptions.`)

	// TODO: add examples
	deleteSubscriptionExample = i18n.T(``)
)

// NewCmdDeleteSubscription creates command responsible for deleting subscriptions.
func NewCmdDeleteSubscription(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription SUBSCRIPTION_NAME",
		Short:   i18n.T("Delete subscription"),
		Long:    deleteSubscriptionLong,
		Example: deleteSubscriptionExample,
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"subscriptions"},
		Run: func(cmd *cobra.Command, args []string) {
			c := eventManagerClient()
			err := deleteSubscription(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	return cmd
}

// CallDeleteSubscription makes the API call to delete an event subscription
func CallDeleteSubscription(c client.EventsClient) ModelAction {
	return func(i interface{}) error {
		subscription := i.(*v1.Subscription)

		deleted, err := c.DeleteSubscription(context.TODO(), "", *subscription.Name)
		if err != nil {
			return err
		}
		*subscription = *deleted
		return nil
	}
}

func deleteSubscription(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EventsClient) error {
	subscriptionModel := v1.Subscription{
		Name: &args[0],
	}
	err := CallDeleteSubscription(c)(&subscriptionModel)
	if err != nil {
		return err
	}
	return formatDeleteSubscriptionOutput(out, false, []*v1.Subscription{&subscriptionModel})
}

func formatDeleteSubscriptionOutput(out io.Writer, list bool, subscriptions []*v1.Subscription) error {
	if w, err := formatOutput(out, list, subscriptions); w {
		return err
	}
	for _, s := range subscriptions {
		_, err := fmt.Fprintf(out, "Deleted subscription: %s\n", *s.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
