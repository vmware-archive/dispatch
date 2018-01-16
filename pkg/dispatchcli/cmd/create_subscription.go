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

	"github.com/vmware/dispatch/pkg/dispatchcli/i18n"
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/subscriptions"
	models "github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

var (
	createSubscriptionLong = i18n.T(`Create dispatch event subscription.`)

	// TODO: add examples
	createSubscriptionExample = i18n.T(``)
	createSubscriptionSecrets []string
)

// NewCmdCreateSubscription creates command responsible for subscription creation.
func NewCmdCreateSubscription(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription TOPIC FUNCTION_NAME",
		Short:   i18n.T("Create subscription"),
		Long:    createSubscriptionLong,
		Example: createSubscriptionExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := createSubscription(out, errOut, cmd, args)
			CheckErr(err)
		},
	}
	cmd.Flags().StringArrayVar(&createSubscriptionSecrets, "secret", []string{}, "Function secrets, can be specified multiple times or a comma-delimited string")

	return cmd
}

func createSubscription(out, errOut io.Writer, cmd *cobra.Command, args []string) error {
	params := &subscriptions.AddSubscriptionParams{
		Context: context.Background(),
		Body: &models.Subscription{
			Topic: &args[0],
			Subscriber: &models.Subscriber{
				// TODO: add support for other types of subscribers
				Type: swag.String("function"),
				Name: &args[1],
			},
			Secrets: createSubscriptionSecrets,
		},
	}
	client := eventManagerClient()
	created, err := client.Subscriptions.AddSubscription(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}
	if dispatchConfig.Json {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "    ")
		return encoder.Encode(*created.Payload)
	}
	fmt.Printf("created subscription: %s\n", created.Payload.Name)
	return nil
}
