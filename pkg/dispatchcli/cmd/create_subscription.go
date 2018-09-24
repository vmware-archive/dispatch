///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"fmt"
	"io"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"github.com/vmware/dispatch/pkg/client"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
)

var (
	createSubscriptionLong = i18n.T(`Create dispatch event subscription.`)

	// TODO: add examples
	createSubscriptionExample   = i18n.T(``)
	createSubscriptionSecrets   []string
	createSubscriptionEventType string
	createSubscriptionName      string
)

// NewCmdCreateSubscription creates command responsible for subscription creation.
func NewCmdCreateSubscription(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription FUNCTION_NAME [--name SUBSCRIPTION_NAME] [--event-type EVENT.TYPE] [--secret SECRET1,SECRET2...]",
		Short:   i18n.T("Create subscription"),
		Long:    createSubscriptionLong,
		Example: createSubscriptionExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := eventManagerClient()
			err := createSubscription(out, errOut, cmd, args, c)
			CheckErr(err)
		},
	}
	cmd.Flags().StringArrayVar(&createSubscriptionSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")

	cmd.Flags().StringVar(&createSubscriptionName, "name", "", "Subscription name. If not specified, will be randomly generated.")
	cmd.Flags().StringVar(&createSubscriptionEventType, "event-type", "", "Event Type to filter on.")

	return cmd
}

// CallCreateSubscription makes the API call to create an event subscription
func CallCreateSubscription(c client.EventsClient) ModelAction {
	return func(i interface{}) error {
		subscription := i.(*v1.Subscription)

		created, err := c.CreateSubscription(context.TODO(), "", subscription)
		if err != nil {
			return err
		}
		*subscription = *created
		return nil
	}
}

func createSubscription(out, errOut io.Writer, cmd *cobra.Command, args []string, c client.EventsClient) error {
	subscription := &v1.Subscription{
		Name:      swag.String(resourceName(createSubscriptionName)),
		EventType: &createSubscriptionEventType,
		Function:  &args[0],
		Secrets:   createSubscriptionSecrets,
	}
	err := CallCreateSubscription(c)(subscription)
	if err != nil {
		return err
	}
	if w, err := formatOutput(out, false, subscription); w {
		return err
	}
	fmt.Printf("created subscription: %s\n", *subscription.Name)
	return nil
}
