///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/swag"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/subscriptions"
)

var (
	createSubscriptionLong = i18n.T(`Create dispatch event subscription.`)

	// TODO: add examples
	createSubscriptionExample    = i18n.T(``)
	createSubscriptionSecrets    []string
	createSubscriptionEventType  string
	createSubscriptionSourceType string
	createSubscriptionName       string
)

// NewCmdCreateSubscription creates command responsible for subscription creation.
func NewCmdCreateSubscription(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription FUNCTION_NAME [--name SUBSCRIPTION_NAME] [--event-type EVENT.TYPE] [--source-type SOURCE-TYPE] [--secret SECRET1,SECRET2...]",
		Short:   i18n.T("Create subscription"),
		Long:    createSubscriptionLong,
		Example: createSubscriptionExample,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := createSubscription(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringArrayVar(&createSubscriptionSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")

	cmd.Flags().StringVar(&createSubscriptionName, "name", "", "Subscription name. If not specified, will be randomly generated.")
	cmd.Flags().StringVar(&createSubscriptionEventType, "event-type", "", "Event Type to filter on.")
	cmd.Flags().StringVar(&createSubscriptionSourceType, "source-type", "dispatch", "Source type to filter on. Most often it will be your event driver type.")

	return cmd
}

// CallCreateSubscription makes the API call to create an event subscription
func CallCreateSubscription(i interface{}) error {
	subscription := i.(*v1.Subscription)
	client := eventManagerClient()
	params := &subscriptions.AddSubscriptionParams{
		Context: context.Background(),
		Body:    subscription,
	}
	created, err := client.Subscriptions.AddSubscription(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	*subscription = *created.Payload
	return nil
}

func createSubscription(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	subscription := &v1.Subscription{
		Name:       swag.String(resourceName(createSubscriptionName)),
		EventType:  &createSubscriptionEventType,
		SourceType: &createSubscriptionSourceType,
		Function:   &args[0],
		Secrets:    createSubscriptionSecrets,
	}
	if cmdFlagApplication != "" {
		subscription.Tags = append(subscription.Tags, &v1.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}
	err := CallCreateSubscription(subscription)
	if err != nil {
		return err
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(subscription)
	}
	fmt.Printf("created subscription: %s\n", *subscription.Name)
	return nil
}
