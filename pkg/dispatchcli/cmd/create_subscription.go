///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/subscriptions"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

var (
	createSubscriptionLong = i18n.T(`Create dispatch event subscription.`)

	// TODO: add examples
	createSubscriptionExample    = i18n.T(``)
	createSubscriptionSecrets    []string
	createSubscriptionEventType  string
	createSubscriptionSourceName string
	createSubscriptionSourceType string
	createSubscriptionName       string
)

// NewCmdCreateSubscription creates command responsible for subscription creation.
func NewCmdCreateSubscription(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription FUNCTION_NAME",
		Short:   i18n.T("Create subscription"),
		Long:    createSubscriptionLong,
		Example: createSubscriptionExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := createSubscription(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringVarP(&cmdFlagApplication, "application", "a", "", "associate with an application")
	cmd.Flags().StringArrayVar(&createSubscriptionSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")

	cmd.Flags().StringVar(&createSubscriptionName, "name", "", "Subscription name")
	cmd.Flags().StringVar(&createSubscriptionEventType, "event-type", "*", "Event Type to filter on")
	cmd.Flags().StringVar(&createSubscriptionSourceName, "source-name", "*", "Source name to filter on. Most often it will be your event driver name.")
	cmd.Flags().StringVar(&createSubscriptionSourceType, "source-type", "*", "Source type to filter on. Most often it will be your event driver type.")

	return cmd
}

func createSubscription(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	params := &subscriptions.AddSubscriptionParams{
		Context: context.Background(),
		Body: &models.Subscription{
			EventType:  &createSubscriptionEventType,
			SourceName: &createSubscriptionSourceName,
			SourceType: &createSubscriptionSourceType,
			Function:   &args[1],
			Secrets:    createSubscriptionSecrets,
			Tags:       []*models.Tag{},
		},
	}
	if cmdFlagApplication != "" {
		params.Body.Tags = append(params.Body.Tags, &models.Tag{
			Key:   "Application",
			Value: cmdFlagApplication,
		})
	}
	client := eventManagerClient()
	created, err := client.Subscriptions.AddSubscription(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if dispatchConfig.JSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Printf("created subscription: %s\n", created.Payload.Name)
	return nil
}
