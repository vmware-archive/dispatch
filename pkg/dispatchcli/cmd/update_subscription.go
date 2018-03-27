///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package cmd

import (
	"github.com/vmware/dispatch/pkg/event-manager/gen/client/subscriptions"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
)

// CallUpdateSubscription makes the API call to update a subscription
func CallUpdateSubscription(input interface{}) error {
	subscription := input.(*models.Subscription)
	params := subscriptions.NewUpdateSubscriptionParams()
	params.SubscriptionName = *subscription.Name
	params.Body = subscription
	_, err := eventManagerClient().Subscriptions.UpdateSubscription(params, GetAuthInfoWriter())
	if err != nil {
		return formatAPIError(err, params)
	}

	return nil
}
