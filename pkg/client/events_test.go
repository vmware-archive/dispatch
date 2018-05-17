///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package client_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/testing/fakeserver"
)

func TestCreateSubscription(t *testing.T) {
	fakeServer := fakeserver.NewFakeServer(nil)
	server := httptest.NewServer(fakeServer)
	defer server.Close()

	eclient := client.NewEventsClient(server.URL, nil, testOrgID)

	subscriptionBody := &v1.Subscription{}

	subscriptionResponse, err := eclient.CreateSubscription(context.Background(), testOrgID, subscriptionBody)
	assert.Error(t, err)
	assert.Nil(t, subscriptionResponse)

	subscriptionMap := toMap(t, subscriptionBody)
	fakeServer.AddResponse("POST", "/v1/event/subscriptions", subscriptionMap, subscriptionMap, 201)
	subscriptionResponse, err = eclient.CreateSubscription(context.Background(), testOrgID, subscriptionBody)
	assert.NoError(t, err)
	assert.Equal(t, subscriptionResponse, subscriptionBody)

}
