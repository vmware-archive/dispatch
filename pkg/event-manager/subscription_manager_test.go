///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/client"
	clientmocks "github.com/vmware/dispatch/pkg/client/mocks"
	"github.com/vmware/dispatch/pkg/events"
	eventsmocks "github.com/vmware/dispatch/pkg/events/mocks"
)

func mockSubscriptionManager(queue events.Transport, fnClient client.FunctionsClient) *subscriptionManager {
	return &subscriptionManager{
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
	fnClient.On("RunFunction", mock.Anything, mock.AnythingOfType("*client.FunctionRun")).Return(&client.FunctionRun{}, nil).Once()
	manager.runFunction("testFunction", ev, []string{"secret1", "secret2"})

	fnClient.On("RunFunction", mock.Anything, mock.AnythingOfType("*client.FunctionRun")).Return(&client.FunctionRun{}, errors.New("testerror")).Once()
	manager.runFunction("testFunction", ev, nil)
	fnClient.AssertNumberOfCalls(t, "RunFunction", 2)
}
