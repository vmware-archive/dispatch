///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	subscription "github.com/vmware/dispatch/pkg/event-manager/gen/client/subscriptions"
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
			err := deleteSubscription(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "filter by application")
	return cmd
}

// CallDeleteSubscription makes the API call to delete an event subscription
func CallDeleteSubscription(i interface{}) error {
	subscriptionModel := i.(*v1.Subscription)
	client := eventManagerClient()
	params := &subscription.DeleteSubscriptionParams{
		Context:          context.Background(),
		SubscriptionName: *subscriptionModel.Name,
		Tags:             []string{},
	}
	deleted, err := client.Subscriptions.DeleteSubscription(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*subscriptionModel = *deleted.Payload
	return nil
}

func deleteSubscription(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	subscriptionModel := v1.Subscription{
		Name: &args[0],
	}
	err := CallDeleteSubscription(&subscriptionModel)
	if err != nil {
		return err
	}
	return formatDeleteSubscriptionOutput(out, false, []*v1.Subscription{&subscriptionModel})
}

func formatDeleteSubscriptionOutput(out io.Writer, list bool, subscriptions []*v1.Subscription) error {
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		if list {
			return encoder.Encode(subscriptions)
		}
		return encoder.Encode(subscriptions[0])
	}
	for _, s := range subscriptions {
		_, err := fmt.Fprintf(out, "Deleted subscription: %s\n", *s.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
