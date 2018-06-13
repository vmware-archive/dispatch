///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package subscriptions

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	clientmocks "github.com/vmware/dispatch/pkg/client/mocks"
	"github.com/vmware/dispatch/pkg/events"
	eventsmocks "github.com/vmware/dispatch/pkg/events/mocks"
)

func mockSubscriptionManager(queue events.Transport, fnClient client.FunctionsClient) *defaultManager {
	return &defaultManager{
		queue:      queue,
		fnClient:   fnClient,
		activeSubs: make(map[string]events.Subscription),
	}
}

func TestRunFunction(t *testing.T) {
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Transport{}
	manager := mockSubscriptionManager(queue, fnClient)
	ev := &events.CloudEvent{}
	fnClient.On("RunFunction", mock.Anything, testOrgID, mock.AnythingOfType("*v1.Run")).Return(&v1.Run{}, nil).Once()
	manager.runFunction(context.Background(), testOrgID, "testFunction", ev, []string{"secret1", "secret2"})

	fnClient.On("RunFunction", mock.Anything, testOrgID, mock.AnythingOfType("*v1.Run")).Return(&v1.Run{}, errors.New("testerror")).Once()
	manager.runFunction(context.Background(), testOrgID, "testFunction", ev, nil)
	fnClient.AssertNumberOfCalls(t, "RunFunction", 2)
}
