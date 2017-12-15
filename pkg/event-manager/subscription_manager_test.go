///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"errors"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/client"
	clientmocks "github.com/vmware/dispatch/pkg/client/mocks"
	"github.com/vmware/dispatch/pkg/events"
	eventsmocks "github.com/vmware/dispatch/pkg/events/mocks"
)

func mockSubscriptionManager(queue events.Queue, fnClient client.FunctionsClient) *subscriptionManager {
	return &subscriptionManager{
		queue:      queue,
		fnClient:   fnClient,
		activeSubs: make(map[string]events.Subscription),
	}
}

func TestRunFunction(t *testing.T) {
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Queue{}
	manager := mockSubscriptionManager(queue, fnClient)
	ev := &events.Event{
		Topic:       "test.topic",
		Body:        []byte("{\"key\": \"value\"}"),
		ContentType: "application/json",
		ID:          uuid.NewV4().String(),
	}
	fnClient.On("RunFunction", mock.Anything, mock.AnythingOfType("*client.FunctionRun")).Return(&client.FunctionRun{}, nil).Once()
	manager.runFunction("testFunction", ev, []string{"secret1", "secret2"})

	fnClient.On("RunFunction", mock.Anything, mock.AnythingOfType("*client.FunctionRun")).Return(&client.FunctionRun{}, errors.New("testerror")).Once()
	manager.runFunction("testFunction", ev, nil)
	fnClient.AssertNumberOfCalls(t, "RunFunction", 2)
}

func TestEmitEvent(t *testing.T) {
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Queue{}
	manager := mockSubscriptionManager(queue, fnClient)
	ev := &events.Event{
		Topic:       "test.topic",
		Body:        []byte("{\"key\": \"value\"}"),
		ContentType: "application/json",
		ID:          uuid.NewV4().String(),
	}
	queue.On("Publish", mock.AnythingOfType("*events.Event")).Return(nil).Once()
	manager.emitEvent("test.topic.new", ev)

	queue.On("Publish", mock.AnythingOfType("*events.Event")).Return(errors.New("testerror")).Once()
	manager.emitEvent("test.topic.new", ev)
	queue.AssertNumberOfCalls(t, "Publish", 2)
}
